package uplink_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/upmaru/terraform-provider-instellar/internal/acceptance"
)

func TestAccUplinkResource_lite(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: buildConfig_lite(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("instellar_uplink.test", "channel_slug", "develop"),
					resource.TestCheckResourceAttr("instellar_uplink.test", "kit_slug", "lite"),
					// Verify computed attribute fields.
					resource.TestCheckResourceAttr("instellar_uplink.test", "current_state", "created"),
					// Dynamic values
					resource.TestCheckResourceAttrSet("instellar_uplink.test", "id"),
					resource.TestCheckResourceAttrSet("instellar_uplink.test", "last_updated"),
				),
			},
			{
				ResourceName:            "instellar_uplink.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"last_updated", "current_state"},
			},
		},
	})
}

func TestAccUplinkResource_pro(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: buildConfig_pro(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("instellar_uplink.test", "channel_slug", "develop"),
					resource.TestCheckResourceAttr("instellar_uplink.test", "kit_slug", "pro"),
					// Verify computed attribute fields.
					resource.TestCheckResourceAttr("instellar_uplink.test", "current_state", "created"),
					// Dynamic values
					resource.TestCheckResourceAttrSet("instellar_uplink.test", "id"),
					resource.TestCheckResourceAttrSet("instellar_uplink.test", "last_updated"),
				),
			},
			{
				ResourceName:            "instellar_uplink.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"last_updated", "current_state"},
			},
		},
	})
}

func buildConfig_lite() string {
	clusterUUID := uuid.New()
	clusterNameSegments := strings.Split(clusterUUID.String(), "-")
	clusterNameSlug := strings.Join([]string{clusterNameSegments[0], clusterNameSegments[1]}, "-")

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
	`
}

func buildConfig_pro() string {
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
			kit_slug = "pro"
			cluster_id = instellar_cluster.test.id
		}
	`
}
