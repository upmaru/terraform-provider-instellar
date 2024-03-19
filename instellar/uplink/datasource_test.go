package uplink_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/upmaru/terraform-provider-instellar/internal/acceptance"
)

func TestAccUplinkDataSource(t *testing.T) {
	clusterUUID := uuid.New()
	clusterNameSegments := strings.Split(clusterUUID.String(), "-")
	clusterNameSlug := strings.Join([]string{clusterNameSegments[0], clusterNameSegments[1]}, "-")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: buildDataConfig(clusterNameSlug, "develop", "lite"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.instellar_uplink.test", "nodes.#", "0"),
				),
			},
		},
	})
}

func buildDataConfig(clusterNameSlug string, channelSlug string, kitSlug string) string {
	return acceptance.ProviderConfig + fmt.Sprintf(`
		resource "instellar_cluster" "test" {
			name = "%s"
			provider_name = "aws"
			region = "ap-southeast-1"
			endpoint = "127.0.0.1:8443"
			password_token = "some-password-or-token"
		}
	`, clusterNameSlug) + fmt.Sprintf(`
		resource "instellar_uplink" "test" {
			channel_slug = "%s"
			kit_slug = "%s"
			cluster_id = instellar_cluster.test.id
		}
	`, channelSlug, kitSlug) + fmt.Sprintf(`
		data "instellar_uplink" "test" {
			id = instellar_uplink.test.id
		}
	`)
}
