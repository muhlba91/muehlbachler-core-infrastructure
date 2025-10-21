import { interpolate } from '@pulumi/pulumi';

import { ServiceAccountData } from '../model/google/service_account_data';

import {
  backupBucketId,
  dnsConfig,
  gcpConfig,
  globalName,
} from './configuration';
import { createIAMMember } from './google/iam/iam_member';
import { createKMSIAMMember } from './google/kms/iam_member';
import { createGCSIAMMember } from './google/storage/iam_member';
import { createGCPServiceAccountAndKey } from './util/google/service_account_user';

/**
 * Creates the Core IAM resources.
 *
 * @returns {ServiceAccountData} the service account
 */
export const createServiceAccount = (): ServiceAccountData => {
  const iam = createGCPServiceAccountAndKey(globalName, gcpConfig.project, {});

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

    createGCSIAMMember(
      backupBucketId,
      `serviceAccount:${email}`,
      'roles/storage.objectAdmin',
    );
    createGCSIAMMember(
      backupBucketId,
      `serviceAccount:${email}`,
      'roles/storage.legacyBucketReader',
    );

    createIAMMember(
      `${email}-dns-admin`,
      interpolate`serviceAccount:${email}`,
      ['roles/dns.admin'],
      { project: dnsConfig.project },
    );
  });

  return iam;
};
