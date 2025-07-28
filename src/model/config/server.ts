/**
 * Defines configuration data for the server.
 */
export interface ServerConfig {
  readonly location: string;
  readonly type: string;
  readonly ipv4: string;
  readonly publicSsh?: boolean;
}
