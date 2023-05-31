package node_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/upmaru/terraform-provider-instellar/internal/acceptance"
)

func TestAccNodeResource(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: buildConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("instellar_node.test", "slug", "pizza-node-ham"),
					// Verify computed attribute fields.
					resource.TestCheckResourceAttr("instellar_node.test", "current_state", "created"),
					// Dynamic values
					resource.TestCheckResourceAttrSet("instellar_node.test", "id"),
					resource.TestCheckResourceAttrSet("instellar_node.test", "last_updated"),
				),
			},
			{
				ResourceName:            "instellar_node.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"last_updated", "current_state"},
			},
		},
	})
}

func buildConfig() string {
	clusterUUID := uuid.New()
	clusterNameSegments := strings.Split(clusterUUID.String(), "-")
	clusterNameSlug := clusterNameSegments[0]

	return acceptance.ProviderConfig + fmt.Sprintf(`
		resource "instellar_cluster" "test" {
			name = "%s"
			provider_name = "aws"
			region = "ap-southeast-1"
			endpoint = "127.0.0.1:8443"
			password_token = "some-password-or-token"
		}
	`, clusterNameSlug) + `
		resource "instellar_uplink" "test" {
			channel_slug = "develop"
			cluster_id = instellar_cluster.test.id
		}
	` + `
		resource "instellar_node" "test" {
			slug = "pizza-node-ham"
			public_ip = "127.0.0.1"
			cluster_id = instellar_cluster.test.id
		}
	`
}
