// Copyright (c) CyberArk
// SPDX-License-Identifier: Apache-2.0

package sia

import (
	"github.com/cyberark/idsec-sdk-golang/pkg/services/sia/settings/actions"
	settingsmodels "github.com/cyberark/idsec-sdk-golang/pkg/services/sia/settings/models"
	tfactions "github.com/cyberark/terraform-provider-idsec/internal/actions"
)

func init() {
	_ = tfactions.Register(tfactions.TerraformServiceConfig{
		ServiceName: "sia-settings",
		Resources: []*tfactions.IdsecServiceTerraformResourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "sia-settings-settings", ActionDescription: "The SIA ListSettings resource.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema: &settingsmodels.IdsecSIASettings{},
				},
				SupportedOperations: []tfactions.IdsecServiceActionOperation{tfactions.CreateOperation, tfactions.ReadOperation, tfactions.UpdateOperation, tfactions.StateOperation},
				ActionsMappings:     map[tfactions.IdsecServiceActionOperation]string{tfactions.CreateOperation: "set-settings", tfactions.ReadOperation: "list-settings", tfactions.UpdateOperation: "set-settings"},
				ImportID:            tfactions.SingletonResourceImportDummyID,
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "sia-settings-adb-mfa-caching", ActionDescription: "The SIA ADB MFA caching resource.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema: &settingsmodels.IdsecSIASettingsAdbMfaCaching{},
				},
				SupportedOperations: []tfactions.IdsecServiceActionOperation{tfactions.CreateOperation, tfactions.ReadOperation, tfactions.UpdateOperation, tfactions.StateOperation},
				ActionsMappings:     map[tfactions.IdsecServiceActionOperation]string{tfactions.CreateOperation: "set-adb-mfa-caching", tfactions.ReadOperation: "adb-mfa-caching", tfactions.UpdateOperation: "set-adb-mfa-caching"},
				ImportID:            tfactions.SingletonResourceImportDummyID,
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "sia-settings-certificate-validation", ActionDescription: "The SIA certificate validation resource.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema: &settingsmodels.IdsecSIASettingsCertificateValidation{},
				},
				SupportedOperations: []tfactions.IdsecServiceActionOperation{tfactions.CreateOperation, tfactions.ReadOperation, tfactions.UpdateOperation, tfactions.StateOperation},
				ActionsMappings:     map[tfactions.IdsecServiceActionOperation]string{tfactions.CreateOperation: "set-certificate-validation", tfactions.ReadOperation: "certificate-validation", tfactions.UpdateOperation: "set-certificate-validation"},
				ImportID:            tfactions.SingletonResourceImportDummyID,
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "sia-settings-k8s-mfa-caching", ActionDescription: "The SIA Kubernetes (K8S) MFA caching resource.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema: &settingsmodels.IdsecSIASettingsK8sMfaCaching{},
				},
				SupportedOperations: []tfactions.IdsecServiceActionOperation{tfactions.CreateOperation, tfactions.ReadOperation, tfactions.UpdateOperation, tfactions.StateOperation},
				ActionsMappings:     map[tfactions.IdsecServiceActionOperation]string{tfactions.CreateOperation: "set-k8s-mfa-caching", tfactions.ReadOperation: "k8s-mfa-caching", tfactions.UpdateOperation: "set-k8s-mfa-caching"},
				ImportID:            tfactions.SingletonResourceImportDummyID,
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "sia-settings-rdp-file-transfer", ActionDescription: "The SIA RDP file transfer resource.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema: &settingsmodels.IdsecSIASettingsRdpFileTransfer{},
				},
				SupportedOperations: []tfactions.IdsecServiceActionOperation{tfactions.CreateOperation, tfactions.ReadOperation, tfactions.UpdateOperation, tfactions.StateOperation},
				ActionsMappings:     map[tfactions.IdsecServiceActionOperation]string{tfactions.CreateOperation: "set-rdp-file-transfer", tfactions.ReadOperation: "rdp-file-transfer", tfactions.UpdateOperation: "set-rdp-file-transfer"},
				ImportID:            tfactions.SingletonResourceImportDummyID,
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "sia-settings-rdp-keyboard-layout", ActionDescription: "The SIA RDP keyboard layout resource.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema: &settingsmodels.IdsecSIASettingsRdpKeyboardLayout{},
				},
				SupportedOperations: []tfactions.IdsecServiceActionOperation{tfactions.CreateOperation, tfactions.ReadOperation, tfactions.UpdateOperation, tfactions.StateOperation},
				ActionsMappings:     map[tfactions.IdsecServiceActionOperation]string{tfactions.CreateOperation: "set-rdp-keyboard-layout", tfactions.ReadOperation: "rdp-keyboard-layout", tfactions.UpdateOperation: "set-rdp-keyboard-layout"},
				ImportID:            tfactions.SingletonResourceImportDummyID,
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "sia-settings-rdp-mfa-caching", ActionDescription: "The SIA RDP MFA caching resource.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema: &settingsmodels.IdsecSIASettingsRdpMfaCaching{},
				},
				SupportedOperations: []tfactions.IdsecServiceActionOperation{tfactions.CreateOperation, tfactions.ReadOperation, tfactions.UpdateOperation, tfactions.StateOperation},
				ActionsMappings:     map[tfactions.IdsecServiceActionOperation]string{tfactions.CreateOperation: "set-rdp-mfa-caching", tfactions.ReadOperation: "rdp-mfa-caching", tfactions.UpdateOperation: "set-rdp-mfa-caching"},
				ImportID:            tfactions.SingletonResourceImportDummyID,
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "sia-settings-rdp-token-mfa-caching", ActionDescription: "The SIA RDP token MFA caching resource.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema: &settingsmodels.IdsecSIASettingsRdpTokenMfaCaching{},
				},
				SupportedOperations: []tfactions.IdsecServiceActionOperation{tfactions.CreateOperation, tfactions.ReadOperation, tfactions.UpdateOperation, tfactions.StateOperation},
				ActionsMappings:     map[tfactions.IdsecServiceActionOperation]string{tfactions.CreateOperation: "set-rdp-token-mfa-caching", tfactions.ReadOperation: "rdp-token-mfa-caching", tfactions.UpdateOperation: "set-rdp-token-mfa-caching"},
				ImportID:            tfactions.SingletonResourceImportDummyID,
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "sia-settings-rdp-recording", ActionDescription: "The SIA RDP recording resource.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema: &settingsmodels.IdsecSIASettingsRdpRecording{},
				},
				SupportedOperations: []tfactions.IdsecServiceActionOperation{tfactions.CreateOperation, tfactions.ReadOperation, tfactions.UpdateOperation, tfactions.StateOperation},
				ActionsMappings:     map[tfactions.IdsecServiceActionOperation]string{tfactions.CreateOperation: "set-rdp-recording", tfactions.ReadOperation: "rdp-recording", tfactions.UpdateOperation: "set-rdp-recording"},
				ImportID:            tfactions.SingletonResourceImportDummyID,
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "sia-settings-ssh-mfa-caching", ActionDescription: "The SIA SSH MFA caching resource.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema: &settingsmodels.IdsecSIASettingsSshMfaCaching{},
				},
				SupportedOperations: []tfactions.IdsecServiceActionOperation{tfactions.CreateOperation, tfactions.ReadOperation, tfactions.UpdateOperation, tfactions.StateOperation},
				ActionsMappings:     map[tfactions.IdsecServiceActionOperation]string{tfactions.CreateOperation: "set-ssh-mfa-caching", tfactions.ReadOperation: "ssh-mfa-caching", tfactions.UpdateOperation: "set-ssh-mfa-caching"},
				ImportID:            tfactions.SingletonResourceImportDummyID,
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "sia-settings-ssh-command-audit", ActionDescription: "The SIA SSH command audit resource.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema: &settingsmodels.IdsecSIASettingsSshCommandAudit{},
				},
				SupportedOperations: []tfactions.IdsecServiceActionOperation{tfactions.CreateOperation, tfactions.ReadOperation, tfactions.UpdateOperation, tfactions.StateOperation},
				ActionsMappings:     map[tfactions.IdsecServiceActionOperation]string{tfactions.CreateOperation: "set-ssh-command-audit", tfactions.ReadOperation: "ssh-command-audit", tfactions.UpdateOperation: "set-ssh-command-audit"},
				ImportID:            tfactions.SingletonResourceImportDummyID,
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "sia-settings-standing-access", ActionDescription: "The SIA standing access resource.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema: &settingsmodels.IdsecSIASettingsStandingAccess{},
				},
				SupportedOperations: []tfactions.IdsecServiceActionOperation{tfactions.CreateOperation, tfactions.ReadOperation, tfactions.UpdateOperation, tfactions.StateOperation},
				ActionsMappings:     map[tfactions.IdsecServiceActionOperation]string{tfactions.CreateOperation: "set-standing-access", tfactions.ReadOperation: "standing-access", tfactions.UpdateOperation: "set-standing-access"},
				ImportID:            tfactions.SingletonResourceImportDummyID,
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "sia-settings-logon-sequence", ActionDescription: "The SIA logon sequence resource.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema: &settingsmodels.IdsecSIASettingsLogonSequence{},
				},
				SupportedOperations: []tfactions.IdsecServiceActionOperation{tfactions.CreateOperation, tfactions.ReadOperation, tfactions.UpdateOperation, tfactions.StateOperation},
				ActionsMappings:     map[tfactions.IdsecServiceActionOperation]string{tfactions.CreateOperation: "set-logon-sequence", tfactions.ReadOperation: "logon-sequence", tfactions.UpdateOperation: "set-logon-sequence"},
				ImportID:            tfactions.SingletonResourceImportDummyID,
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "sia-settings-self-hosted-pam", ActionDescription: "The SIA PAM Self-Hosted resource.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema: &settingsmodels.IdsecSIASettingsSelfHostedPam{},
				},
				SupportedOperations: []tfactions.IdsecServiceActionOperation{tfactions.CreateOperation, tfactions.ReadOperation, tfactions.UpdateOperation, tfactions.StateOperation},
				ActionsMappings:     map[tfactions.IdsecServiceActionOperation]string{tfactions.CreateOperation: "set-self-hosted-pam", tfactions.ReadOperation: "self-hosted-pam", tfactions.UpdateOperation: "set-self-hosted-pam"},
				ImportID:            tfactions.SingletonResourceImportDummyID,
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "sia-settings-rdp-kerberos-auth-mode", ActionDescription: "The SIA RDP Kerberos auth mode resource.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema: &settingsmodels.IdsecSIASettingsRdpKerberosAuthMode{},
				},
				SupportedOperations: []tfactions.IdsecServiceActionOperation{tfactions.CreateOperation, tfactions.ReadOperation, tfactions.UpdateOperation, tfactions.StateOperation},
				ActionsMappings:     map[tfactions.IdsecServiceActionOperation]string{tfactions.CreateOperation: "set-rdp-kerberos-auth-mode", tfactions.ReadOperation: "rdp-kerberos-auth-mode", tfactions.UpdateOperation: "set-rdp-kerberos-auth-mode"},
				ImportID:            tfactions.SingletonResourceImportDummyID,
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "sia-settings-rdp-transcription", ActionDescription: "The SIA RDP transcription resource.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema: &settingsmodels.IdsecSIASettingsRdpTranscription{},
				},
				SupportedOperations: []tfactions.IdsecServiceActionOperation{tfactions.CreateOperation, tfactions.ReadOperation, tfactions.UpdateOperation, tfactions.StateOperation},
				ActionsMappings:     map[tfactions.IdsecServiceActionOperation]string{tfactions.CreateOperation: "set-rdp-transcription", tfactions.ReadOperation: "rdp-transcription", tfactions.UpdateOperation: "set-rdp-transcription"},
				ImportID:            tfactions.SingletonResourceImportDummyID,
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "sia-settings-ssh-recording", ActionDescription: "The SIA SSH recording resource.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema: &settingsmodels.IdsecSIASettingsSshRecording{},
				},
				SupportedOperations: []tfactions.IdsecServiceActionOperation{tfactions.CreateOperation, tfactions.ReadOperation, tfactions.UpdateOperation, tfactions.StateOperation},
				ActionsMappings:     map[tfactions.IdsecServiceActionOperation]string{tfactions.CreateOperation: "set-ssh-recording", tfactions.ReadOperation: "ssh-recording", tfactions.UpdateOperation: "set-ssh-recording"},
				ImportID:            tfactions.SingletonResourceImportDummyID,
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "sia-settings-validate-fingerprint-for-ssh-zero-standing", ActionDescription: "The SIA SSH fingerprint validation for Zero Standing connections resource.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema: &settingsmodels.IdsecSIASettingsValidateFingerprintForSSHZeroStanding{},
				},
				SupportedOperations: []tfactions.IdsecServiceActionOperation{tfactions.CreateOperation, tfactions.ReadOperation, tfactions.UpdateOperation, tfactions.StateOperation},
				ActionsMappings:     map[tfactions.IdsecServiceActionOperation]string{tfactions.CreateOperation: "set-validate-fingerprint-for-ssh-zero-standing", tfactions.ReadOperation: "validate-fingerprint-for-ssh-zero-standing", tfactions.UpdateOperation: "set-validate-fingerprint-for-ssh-zero-standing"},
				ImportID:            tfactions.SingletonResourceImportDummyID,
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "sia-settings-rdp-file-parameters", ActionDescription: "The SIA RDP File Parameters resource.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema: &settingsmodels.IdsecSIASettingsRdpFileParameters{},
				},
				SupportedOperations: []tfactions.IdsecServiceActionOperation{tfactions.CreateOperation, tfactions.ReadOperation, tfactions.UpdateOperation, tfactions.StateOperation},
				ActionsMappings:     map[tfactions.IdsecServiceActionOperation]string{tfactions.CreateOperation: "set-rdp-file-parameters", tfactions.ReadOperation: "rdp-file-parameters", tfactions.UpdateOperation: "set-rdp-file-parameters"},
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "sia-settings-rdp-file-signing", ActionDescription: "The SIA RDP File Signing resource.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema: &settingsmodels.IdsecSIASettingsRdpFileSigning{},
				},
				SupportedOperations: []tfactions.IdsecServiceActionOperation{tfactions.CreateOperation, tfactions.ReadOperation, tfactions.UpdateOperation, tfactions.StateOperation},
				ActionsMappings:     map[tfactions.IdsecServiceActionOperation]string{tfactions.CreateOperation: "set-rdp-file-signing", tfactions.ReadOperation: "rdp-file-signing", tfactions.UpdateOperation: "set-rdp-file-signing"},
				ImportID:            tfactions.SingletonResourceImportDummyID,
			},
		},
		DataSources: []*tfactions.IdsecServiceTerraformDataSourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "sia-settings-settings", ActionDescription: "The SIA ListSettings data source.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema: &settingsmodels.IdsecSIASettings{},
				},
				DataSourceAction: "list-settings",
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "sia-settings-adb-mfa-caching", ActionDescription: "The SIA ADB MFA caching settings data source.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema: &settingsmodels.IdsecSIASettingsAdbMfaCaching{},
				},
				DataSourceAction: "adb-mfa-caching",
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "sia-settings-certificate-validation", ActionDescription: "The SIA certificate validation settings data source.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema: &settingsmodels.IdsecSIASettingsCertificateValidation{},
				},
				DataSourceAction: "certificate-validation",
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "sia-settings-k8s-mfa-caching", ActionDescription: "The SIA Kubernetes (K8s) MFA caching settings data source.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema: &settingsmodels.IdsecSIASettingsK8sMfaCaching{},
				},
				DataSourceAction: "k8s-mfa-caching",
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "sia-settings-rdp-file-transfer", ActionDescription: "The SIA RDP file transfer settings data source.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema: &settingsmodels.IdsecSIASettingsRdpFileTransfer{},
				},
				DataSourceAction: "rdp-file-transfer",
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "sia-settings-rdp-keyboard-layout", ActionDescription: "The SIA RDP keyboard layout settings data source.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema: &settingsmodels.IdsecSIASettingsRdpKeyboardLayout{},
				},
				DataSourceAction: "rdp-keyboard-layout",
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "sia-settings-rdp-mfa-caching", ActionDescription: "The SIA RDP MFA caching settings data source.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema: &settingsmodels.IdsecSIASettingsRdpMfaCaching{},
				},
				DataSourceAction: "rdp-mfa-caching",
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "sia-settings-rdp-token-mfa-caching", ActionDescription: "The SIA RDP token MFA caching settings data source.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema: &settingsmodels.IdsecSIASettingsRdpTokenMfaCaching{},
				},
				DataSourceAction: "rdp-token-mfa-caching",
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "sia-settings-rdp-recording", ActionDescription: "The SIA RDP recording settings data source.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema: &settingsmodels.IdsecSIASettingsRdpRecording{},
				},
				DataSourceAction: "rdp-recording",
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "sia-settings-ssh-mfa-caching", ActionDescription: "The SIA SSH MFA caching settings data source.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema: &settingsmodels.IdsecSIASettingsSshMfaCaching{},
				},
				DataSourceAction: "ssh-mfa-caching",
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "sia-settings-ssh-command-audit", ActionDescription: "The SIA SSH command audit settings data source.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema: &settingsmodels.IdsecSIASettingsSshCommandAudit{},
				},
				DataSourceAction: "ssh-command-audit",
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "sia-settings-standing-access", ActionDescription: "The SIA standing access settings data source.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema: &settingsmodels.IdsecSIASettingsStandingAccess{},
				},
				DataSourceAction: "standing-access",
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "sia-settings-logon-sequence", ActionDescription: "The SIA logon sequence settings data source.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema: &settingsmodels.IdsecSIASettingsLogonSequence{},
				},
				DataSourceAction: "logon-sequence",
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "sia-settings-self-hosted-pam", ActionDescription: "The SIA PAM Self-Hosted settings data source.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema: &settingsmodels.IdsecSIASettingsSelfHostedPam{},
				},
				DataSourceAction: "self-hosted-pam",
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "sia-settings-rdp-kerberos-auth-mode", ActionDescription: "The SIA RDP Kerberos auth mode settings data source.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema: &settingsmodels.IdsecSIASettingsRdpKerberosAuthMode{},
				},
				DataSourceAction: "rdp-kerberos-auth-mode",
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "sia-settings-rdp-transcription", ActionDescription: "The SIA RDP transcription settings data source.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema: &settingsmodels.IdsecSIASettingsRdpTranscription{},
				},
				DataSourceAction: "rdp-transcription",
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "sia-settings-ssh-recording", ActionDescription: "The SIA SSH recording settings data source.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema: &settingsmodels.IdsecSIASettingsSshRecording{},
				},
				DataSourceAction: "ssh-recording",
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "sia-settings-validate-fingerprint-for-ssh-zero-standing", ActionDescription: "The SIA SSH fingerprint validation for Zero Standing connections data source.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema: &settingsmodels.IdsecSIASettingsValidateFingerprintForSSHZeroStanding{},
				},
				DataSourceAction: "validate-fingerprint-for-ssh-zero-standing",
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "sia-settings-rdp-file-parameters", ActionDescription: "The SIA RDP File Parameters settings data source.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema: &settingsmodels.IdsecSIASettingsRdpFileParameters{},
				},
				DataSourceAction: "rdp-file-parameters",
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "sia-settings-rdp-file-signing", ActionDescription: "The SIA RDP File Signing settings data source.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					StateSchema: &settingsmodels.IdsecSIASettingsRdpFileSigning{},
				},
				DataSourceAction: "rdp-file-signing",
			},
		},
	})
}
