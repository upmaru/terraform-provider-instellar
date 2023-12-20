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
	clusterUUID := uuid.New()
	clusterNameSegments := strings.Split(clusterUUID.String(), "-")
	clusterNameSlug := strings.Join([]string{clusterNameSegments[0], clusterNameSegments[1]}, "-")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: buildConfig(clusterNameSlug, "127.0.0.1"),
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
			{
				Config: buildConfig(clusterNameSlug, "38.56.93.42"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("instellar_node.test", "public_ip", "38.56.93.42"),
					// Verify computed attribute fields.
					resource.TestCheckResourceAttr("instellar_node.test", "current_state", "syncing"),
					// Dynamic values
					resource.TestCheckResourceAttrSet("instellar_node.test", "id"),
					resource.TestCheckResourceAttrSet("instellar_node.test", "last_updated"),
				),
			},
		},
	})
}

func buildConfig(clusterNameSlug string, publicIp string) string {
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
			kit_slug = "lite"
			cluster_id = instellar_cluster.test.id
		}
	` + fmt.Sprintf(`
		resource "instellar_node" "test" {
			slug = "pizza-node-ham"
			public_ip = "%s"
			cluster_id = instellar_cluster.test.id
		}
	`, publicIp)
}
