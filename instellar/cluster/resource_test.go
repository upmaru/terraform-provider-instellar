package cluster_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/upmaru/terraform-provider-instellar/instellar/acceptance"
)

func TestAccClusterResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: acceptance.ProviderConfig + `
					resource "instellar_cluster" "test" {
						name = "example-cluster"
						provider = "aws"
						region = "ap-southeast-1"
						password_token = "some-password-or-token"
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("instellar_cluster.test", "name", "example-cluster"),
					// Verify computed attribute fields.
					resource.TestCheckResourceAttr("instellar_cluster.test", "slug", "example-cluster"),
					resource.TestCheckResourceAttr("instellar_cluster.test", "current_state", "syncing"),
					// Verify dynamic vlaues have value set
					resource.TestCheckResourceAttrSet("instellar_cluster.test", "id"),
					resource.TestCheckResourceAttrSet("instellar_cluster.test", "last_updated"),
				),
			},
			{
				ResourceName:            "instellar_cluster.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"last_updated"},
			},
		},
	})
}
