package netbird

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/muhlba91/pulumi-shared-library/pkg/util/file"
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	rModel "github.com/muhlba91/pulumi-shared-library/pkg/model/rotation"
	"github.com/muhlba91/pulumi-shared-library/pkg/util/rotation"

	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/model/netbird"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/util/remotefile"
	"github.com/muhlba91/muehlbachler-core-infrastructure/pkg/util/script"
)

const (
	// patExpiryDays is the lifetime of the personal access token in days (maximum allowed by NetBird).
	patExpiryDays = 365
	// patRotateBeforeDays is the remaining lifetime in days at which the token is rotated.
	patRotateBeforeDays = 90
	// patCheckIntervalDays is the interval in days in which the token is checked for rotation.
	patCheckIntervalDays = 30
)

// initialize sets up NetBird (initial owner and personal access token) on the remote server via SSH.
// The script is idempotent: the token is persisted on the server (and backed up), and only created once.
// It is re-run periodically (rotation trigger) and rotates the token if it expires in less than patRotateBeforeDays.
// ctx: Pulumi context.
// sshIPv4: The IPv4 address of the server to connect to via SSH.
// privateKeyPem: The private key in PEM format to use for SSH authentication.
// netbirdData: NetBird configuration data.
// dependsOn: Pulumi resource option to specify dependencies.
func initialize(
	ctx *pulumi.Context,
	sshIPv4 pulumi.StringOutput,
	privateKeyPem pulumi.StringOutput,
	netbirdData *netbird.Data,
	dependsOn pulumi.ResourceOrInvokeOption,
) (pulumi.StringOutput, error) {
	conn := &remote.ConnectionArgs{
		Host:       sshIPv4,
		PrivateKey: privateKeyPem,
		User:       pulumi.String("root"),
	}

	// the request body of the setup contains the password: it is removed by the script once used
	setupBody, _ := netbirdData.Admin.Password.ApplyT(func(password string) (string, error) {
		body, err := json.Marshal(map[string]any{
			"name":          netbirdData.Admin.Name,
			"email":         netbirdData.Admin.Email,
			"password":      password, //nolint:goconst // template key
			"create_pat":    true,
			"pat_expire_in": patExpiryDays,
		})
		return string(body), err
	}).(pulumi.StringOutput)
	_, setupBodyCopy := remotefile.CopyRendered(
		ctx,
		"remote-copy-netbird-setup",
		"./outputs/netbird_setup.json",
		pulumi.String("/opt/netbird/setup.json"),
		setupBody,
		conn,
		dependsOn,
	)

	initScript, sErr := file.ReadContents("./assets/netbird/init.sh")
	if sErr != nil {
		return pulumi.StringOutput{}, sErr
	}

	// the settings are passed as variables prefixed to the script: the environment of a command requires sshd to accept them
	shebang, scriptBody, _ := strings.Cut(initScript, "\n")
	initScript = fmt.Sprintf(
		"%s\nexport EXPIRY_DAYS=%d\nexport ROTATE_BEFORE_DAYS=%d\n%s",
		shebang,
		patExpiryDays,
		patRotateBeforeDays,
		scriptBody,
	)

	rotation, rErr := rotationTrigger(ctx)
	if rErr != nil {
		return pulumi.StringOutput{}, rErr
	}

	cmd, cErr := remote.NewCommand(ctx, "netbird-init", &remote.CommandArgs{
		Create:     pulumi.StringPtr(initScript),
		Update:     pulumi.StringPtr(initScript),
		Triggers:   pulumi.Array{rotation},
		Connection: conn,
	},
		dependsOn,
		remotefile.DependsOnAll(setupBodyCopy),
		pulumi.AdditionalSecretOutputs([]string{"stdout", "stderr"}),
		pulumi.Timeouts(&pulumi.CustomTimeouts{
			Create: "40m",
			Update: "40m",
		}))
	if cErr != nil {
		return pulumi.StringOutput{}, cErr
	}

	token, _ := cmd.Stdout.ApplyT(func(stdout string) (string, error) {
		parsedTokens, err := script.ParseTokens(stdout)
		if err != nil {
			return "", err
		}

		// the token file also contains the user and token identifiers, and the expiration: only the token is needed
		token, _ := parsedTokens["token"].(string)
		if token == "" {
			return "", errors.New("netbird: the init output does not contain a token")
		}

		return token, nil
	}).(pulumi.StringOutput)

	return token, nil
}

// rotationTrigger creates the schedule re-running the initialization to check, and rotate the token.
// ctx: Pulumi context.
func rotationTrigger(ctx *pulumi.Context) (pulumi.StringOutput, error) {
	trigger, err := rotation.Trigger(ctx, "netbird-pat", &rModel.Options{Days: patCheckIntervalDays})
	if err != nil {
		return pulumi.StringOutput{}, err
	}
	return *trigger, nil
}
