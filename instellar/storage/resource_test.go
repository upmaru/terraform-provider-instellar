package storage_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/upmaru/terraform-provider-instellar/internal/acceptance"
)

func TestAccStorageResource(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: buildConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("instellar_storage.test", "host", "s3.amazonaws.com"),
					// Verify computed attribute fields.
					resource.TestCheckResourceAttr("instellar_storage.test", "current_state", "syncing"),
					// Dynamic values
					resource.TestCheckResourceAttrSet("instellar_storage.test", "id"),
					resource.TestCheckResourceAttrSet("instellar_storage.test", "last_updated"),
				),
			},
			{
				ResourceName:            "instellar_storage.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"last_updated", "current_state"},
			},
		},
	})
}

func buildConfig() string {
	return acceptance.ProviderConfig + `
		resource "instellar_storage" "test" {
		  host = "s3.amazonaws.com"
		  bucket = "mybucket"
			region = "ap-southeast-1"
			access_key_id = "somekey"
			secret_access_key = "somesecret"
		}
	`
}
