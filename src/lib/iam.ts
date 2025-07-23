import { interpolate } from '@pulumi/pulumi';

import { ServiceAccountData } from '../model/google/service_account_data';

import { dnsConfig, gcpConfig, globalName } from './configuration';
import { createIAMMember } from './google/iam/iam_member';
import { createKMSIAMMember } from './google/kms/iam_member';
import { createGCPServiceAccountAndKey } from './util/google/service_account_user';

/**
 * Creates the Hashicorp Vault IAM resources.
 *
 * @returns {ServiceAccountData} the service account
 */
export const createServiceAccount = (): ServiceAccountData => {
  const iam = createGCPServiceAccountAndKey('vault', gcpConfig.project, {});

  iam.serviceAccount.email.apply((email) => {
    createKMSIAMMember(
      `${gcpConfig.project}/${gcpConfig.encryptionKey.location}/${gcpConfig.encryptionKey.keyringId}`,
      `serviceAccount:${email}`,
      'roles/cloudkms.cryptoKeyEncrypterDecrypter',
    );
    createKMSIAMMember(
      `${gcpConfig.project}/${gcpConfig.encryptionKey.location}/${gcpConfig.encryptionKey.keyringId}`,
      `serviceAccount:${email}`,
      'roles/cloudkms.signerVerifier',
    );
    createKMSIAMMember(
      `${gcpConfig.project}/${gcpConfig.encryptionKey.location}/${gcpConfig.encryptionKey.keyringId}`,
      `serviceAccount:${email}`,
      'roles/cloudkms.viewer',
    );

    createIAMMember(
      `${globalName}-dns-admin`,
      interpolate`serviceAccount:${iam.serviceAccount.email}`,
      ['roles/dns.admin'],
      { project: dnsConfig.project },
    );
  });

  return iam;
};
