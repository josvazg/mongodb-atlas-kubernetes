package atlasproject

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/atlas/mongodbatlas"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/mongodb/mongodb-atlas-kubernetes/v2/internal/pointer"
	akov2 "github.com/mongodb/mongodb-atlas-kubernetes/v2/pkg/api/v1"
	"github.com/mongodb/mongodb-atlas-kubernetes/v2/pkg/api/v1/common"
	"github.com/mongodb/mongodb-atlas-kubernetes/v2/pkg/controller/workflow"
)

func TestReadEncryptionAtRestSecrets(t *testing.T) {
	t.Run("AWS with correct secret data", func(t *testing.T) {
		secretData := map[string][]byte{
			"AccessKeyID":         []byte("testAccessKeyID"),
			"SecretAccessKey":     []byte("testSecretAccessKey"),
			"CustomerMasterKeyID": []byte("testCustomerMasterKeyID"),
			"RoleID":              []byte("testRoleID"),
		}

		kk := fake.NewClientBuilder().WithRuntimeObjects([]runtime.Object{
			&v1.Secret{
				Data: secretData,
				TypeMeta: metav1.TypeMeta{
					Kind:       "Secret",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "aws-secret",
					Namespace: "test",
				},
			},
		}...).Build()

		service := &workflow.Context{}

		encRest := &akov2.EncryptionAtRest{
			AwsKms: akov2.AwsKms{
				Enabled: pointer.MakePtr(true),
				SecretRef: common.ResourceRefNamespaced{
					Name:      "aws-secret",
					Namespace: "test",
				},
				Region: "testRegion",
			},
		}

		err := readEncryptionAtRestSecrets(kk, service, encRest, "test")
		assert.Nil(t, err)

		assert.Equal(t, string(secretData["CustomerMasterKeyID"]), encRest.AwsKms.CustomerMasterKeyID())
		assert.Equal(t, string(secretData["RoleID"]), encRest.AwsKms.RoleID())
	})

	t.Run("AWS with correct secret data (fallback namespace)", func(t *testing.T) {
		secretData := map[string][]byte{
			"AccessKeyID":         []byte("testAccessKeyID"),
			"SecretAccessKey":     []byte("testSecretAccessKey"),
			"CustomerMasterKeyID": []byte("testCustomerMasterKeyID"),
			"RoleID":              []byte("testRoleID"),
		}

		kk := fake.NewClientBuilder().WithRuntimeObjects([]runtime.Object{
			&v1.Secret{
				Data: secretData,
				TypeMeta: metav1.TypeMeta{
					Kind:       "Secret",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "aws-secret",
					Namespace: "test-fallback-ns",
				},
			},
		}...).Build()

		service := &workflow.Context{}

		encRest := &akov2.EncryptionAtRest{
			AwsKms: akov2.AwsKms{
				Enabled: pointer.MakePtr(true),
				SecretRef: common.ResourceRefNamespaced{
					Name: "aws-secret",
				},
			},
		}

		err := readEncryptionAtRestSecrets(kk, service, encRest, "test-fallback-ns")
		assert.Nil(t, err)

		assert.Equal(t, string(secretData["CustomerMasterKeyID"]), encRest.AwsKms.CustomerMasterKeyID())
		assert.Equal(t, string(secretData["RoleID"]), encRest.AwsKms.RoleID())
	})

	t.Run("AWS with missing fields", func(t *testing.T) {
		secretData := map[string][]byte{
			"AccessKeyID":         []byte("testKeyID"),
			"SecretAccessKey":     []byte("testSecretAccessKey"),
			"CustomerMasterKeyID": []byte("testCustomerMasterKeyID"),
		}

		kk := fake.NewClientBuilder().WithRuntimeObjects([]runtime.Object{
			&v1.Secret{
				Data: secretData,
				TypeMeta: metav1.TypeMeta{
					Kind:       "Secret",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "aws-secret",
					Namespace: "test",
				},
			},
		}...).Build()

		service := &workflow.Context{}

		encRest := &akov2.EncryptionAtRest{
			AwsKms: akov2.AwsKms{
				Enabled: pointer.MakePtr(true),
				SecretRef: common.ResourceRefNamespaced{
					Name:      "aws-secret",
					Namespace: "test",
				},
			},
		}

		err := readEncryptionAtRestSecrets(kk, service, encRest, "test")
		assert.NotNil(t, err)
	})

	t.Run("GCP with correct secret data", func(t *testing.T) {
		secretData := map[string][]byte{
			"ServiceAccountKey":    []byte("testServiceAccountKey"),
			"KeyVersionResourceID": []byte("testKeyVersionResourceID"),
		}

		kk := fake.NewClientBuilder().WithRuntimeObjects([]runtime.Object{
			&v1.Secret{
				Data: secretData,
				TypeMeta: metav1.TypeMeta{
					Kind:       "Secret",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "gcp-secret",
					Namespace: "test",
				},
			},
		}...).Build()

		service := &workflow.Context{}

		encRest := &akov2.EncryptionAtRest{
			GoogleCloudKms: akov2.GoogleCloudKms{
				Enabled: pointer.MakePtr(true),
				SecretRef: common.ResourceRefNamespaced{
					Name: "gcp-secret",
				},
			},
		}

		err := readEncryptionAtRestSecrets(kk, service, encRest, "test")
		assert.Nil(t, err)

		assert.Equal(t, string(secretData["ServiceAccountKey"]), encRest.GoogleCloudKms.ServiceAccountKey())
		assert.Equal(t, string(secretData["KeyVersionResourceID"]), encRest.GoogleCloudKms.KeyVersionResourceID())
	})

	t.Run("GCP with missing fields", func(t *testing.T) {
		secretData := map[string][]byte{
			"ServiceAccountKey": []byte("testServiceAccountKey"),
		}

		kk := fake.NewClientBuilder().WithRuntimeObjects([]runtime.Object{
			&v1.Secret{
				Data: secretData,
				TypeMeta: metav1.TypeMeta{
					Kind:       "Secret",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "gcp-secret",
					Namespace: "test",
				},
			},
		}...).Build()

		service := &workflow.Context{}

		encRest := &akov2.EncryptionAtRest{
			GoogleCloudKms: akov2.GoogleCloudKms{
				Enabled: pointer.MakePtr(true),
				SecretRef: common.ResourceRefNamespaced{
					Name: "gcp-secret",
				},
			},
		}

		err := readEncryptionAtRestSecrets(kk, service, encRest, "test")
		assert.NotNil(t, err)
	})

	t.Run("Azure with correct secret data", func(t *testing.T) {
		secretData := map[string][]byte{
			"Secret":         []byte("testClientSecret"),
			"SubscriptionID": []byte("testSubscriptionID"),
			"KeyVaultName":   []byte("testKeyVaultName"),
			"KeyIdentifier":  []byte("testKeyIdentifier"),
		}

		kk := fake.NewClientBuilder().WithRuntimeObjects([]runtime.Object{
			&v1.Secret{
				Data: secretData,
				TypeMeta: metav1.TypeMeta{
					Kind:       "Secret",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "azure-secret",
					Namespace: "test",
				},
			},
		}...).Build()

		service := &workflow.Context{}

		encRest := &akov2.EncryptionAtRest{
			AzureKeyVault: akov2.AzureKeyVault{
				Enabled: pointer.MakePtr(true),
				SecretRef: common.ResourceRefNamespaced{
					Name: "azure-secret",
				},
			},
		}

		err := readEncryptionAtRestSecrets(kk, service, encRest, "test")
		assert.Nil(t, err)

		assert.Equal(t, string(secretData["Secret"]), encRest.AzureKeyVault.Secret())
		assert.Equal(t, string(secretData["SubscriptionID"]), encRest.AzureKeyVault.SubscriptionID())
		assert.Equal(t, string(secretData["KeyVaultName"]), encRest.AzureKeyVault.KeyVaultName())
		assert.Equal(t, string(secretData["KeyIdentifier"]), encRest.AzureKeyVault.KeyIdentifier())
	})

	t.Run("Azure with missing fields", func(t *testing.T) {
		secretData := map[string][]byte{
			"ClientID":          []byte("testClientID"),
			"AzureEnvironment":  []byte("testAzureEnvironment"),
			"SubscriptionID":    []byte("testSubscriptionID"),
			"ResourceGroupName": []byte("testResourceGroupName"),
		}

		kk := fake.NewClientBuilder().WithRuntimeObjects([]runtime.Object{
			&v1.Secret{
				Data: secretData,
				TypeMeta: metav1.TypeMeta{
					Kind:       "Secret",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "gcp-secret",
					Namespace: "test",
				},
			},
		}...).Build()

		service := &workflow.Context{}

		encRest := &akov2.EncryptionAtRest{
			AzureKeyVault: akov2.AzureKeyVault{
				Enabled: pointer.MakePtr(true),
				SecretRef: common.ResourceRefNamespaced{
					Name: "gcp-secret",
				},
			},
		}

		err := readEncryptionAtRestSecrets(kk, service, encRest, "test")
		assert.NotNil(t, err)
	})
}

func TestIsEncryptionAtlasEmpty(t *testing.T) {
	spec := &akov2.EncryptionAtRest{}
	isEmpty := IsEncryptionSpecEmpty(spec)
	assert.True(t, isEmpty, "Empty spec should be empty")

	spec.AwsKms.Enabled = pointer.MakePtr(true)
	isEmpty = IsEncryptionSpecEmpty(spec)
	assert.False(t, isEmpty, "Non-empty spec")

	spec.AwsKms.Enabled = pointer.MakePtr(false)
	isEmpty = IsEncryptionSpecEmpty(spec)
	assert.True(t, isEmpty, "Enabled flag set to false is same as empty")
}

func TestAtlasInSync(t *testing.T) {
	areInSync, err := AtlasInSync(nil, nil)
	assert.NoError(t, err)
	assert.True(t, areInSync, "Both atlas and spec are nil")

	groupID := "0"
	atlas := mongodbatlas.EncryptionAtRest{
		GroupID: groupID,
		AwsKms: mongodbatlas.AwsKms{
			Enabled: pointer.MakePtr(true),
		},
	}
	spec := akov2.EncryptionAtRest{
		AwsKms: akov2.AwsKms{
			Enabled: pointer.MakePtr(true),
		},
	}

	areInSync, err = AtlasInSync(nil, &spec)
	assert.NoError(t, err)
	assert.False(t, areInSync, "Nil atlas")

	areInSync, err = AtlasInSync(&atlas, nil)
	assert.NoError(t, err)
	assert.False(t, areInSync, "Nil spec")

	areInSync, err = AtlasInSync(&atlas, &spec)
	assert.NoError(t, err)
	assert.True(t, areInSync, "Both are the same")

	spec.AwsKms.Enabled = pointer.MakePtr(false)
	areInSync, err = AtlasInSync(&atlas, &spec)
	assert.NoError(t, err)
	assert.False(t, areInSync, "Atlas is disabled")

	atlas.AwsKms.Enabled = pointer.MakePtr(false)
	areInSync, err = AtlasInSync(&atlas, &spec)
	assert.NoError(t, err)
	assert.True(t, areInSync, "Both are disabled")

	atlas.AwsKms.RoleID = "example"
	areInSync, err = AtlasInSync(&atlas, &spec)
	assert.NoError(t, err)
	assert.True(t, areInSync, "Both are disabled but atlas RoleID field")

	spec.AwsKms.Enabled = pointer.MakePtr(true)
	areInSync, err = AtlasInSync(&atlas, &spec)
	assert.NoError(t, err)
	assert.False(t, areInSync, "Spec is re-enabled")

	atlas.AwsKms.Enabled = pointer.MakePtr(true)
	areInSync, err = AtlasInSync(&atlas, &spec)
	assert.NoError(t, err)
	assert.True(t, areInSync, "Both are re-enabled and only RoleID is different")

	atlas = mongodbatlas.EncryptionAtRest{
		AwsKms: mongodbatlas.AwsKms{
			Enabled:             pointer.MakePtr(true),
			CustomerMasterKeyID: "testCustomerMasterKeyID",
			Region:              "US_EAST_1",
			RoleID:              "testRoleID",
			Valid:               pointer.MakePtr(true),
		},
		AzureKeyVault: mongodbatlas.AzureKeyVault{
			Enabled: pointer.MakePtr(false),
		},
		GoogleCloudKms: mongodbatlas.GoogleCloudKms{
			Enabled: pointer.MakePtr(false),
		},
	}
	spec = akov2.EncryptionAtRest{
		AwsKms: akov2.AwsKms{
			Enabled: pointer.MakePtr(true),
			Region:  "US_EAST_1",
			Valid:   pointer.MakePtr(true),
		},
		AzureKeyVault:  akov2.AzureKeyVault{},
		GoogleCloudKms: akov2.GoogleCloudKms{},
	}
	spec.AwsKms.SetSecrets("testCustomerMasterKeyID", "testRoleID")

	areInSync, err = AtlasInSync(&atlas, &spec)
	assert.NoError(t, err)
	assert.True(t, areInSync, "Realistic example. should be equal")
}
