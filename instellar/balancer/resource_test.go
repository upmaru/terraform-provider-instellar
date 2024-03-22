package balancer_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/upmaru/terraform-provider-instellar/internal/acceptance"
)

func TestAccBalancerResource(t *testing.T) {
	clusterUUID := uuid.New()
	clusterNameSegments := strings.Split(clusterUUID.String(), "-")
	clusterNameSlug := strings.Join([]string{clusterNameSegments[0], clusterNameSegments[1]}, "-")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: buildConfig(clusterNameSlug),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("instellar_balancer.test", "name", "test-balancer"),
					resource.TestCheckResourceAttr("instellar_balancer.test", "address", "some.address.com"),
					// Verify computed attribute fields.
					resource.TestCheckResourceAttr("instellar_balancer.test", "current_state", "active"),
					// Dynamic values
					resource.TestCheckResourceAttrSet("instellar_balancer.test", "id"),
					resource.TestCheckResourceAttrSet("instellar_balancer.test", "last_updated"),
				),
			},
			{
				ResourceName:            "instellar_balancer.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"last_updated", "current_state"},
			},
		},
	})
}

func buildConfig(clusterName string) string {
	return acceptance.ProviderConfig + fmt.Sprintf(`
		resource "instellar_cluster" "test" {
			name = "%s"
			provider_name = "aws"
			region = "ap-southeast-1"
			endpoint = "127.0.0.1:8443"
			password_token = "some-password-or-token"
		}

		resource "instellar_balancer" "test" {
			name = "test-balancer"
			address = "some.address.com"
			cluster_id = instellar_cluster.test.id
		}
	`, clusterName)
}