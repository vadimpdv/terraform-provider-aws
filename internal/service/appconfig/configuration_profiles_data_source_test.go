package appconfig_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/appconfig"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccAppConfigConfigurationProfilesDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	appName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_appconfig_configuration_profiles.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, appconfig.EndpointsID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, appconfig.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationProfilesDataSourceConfig_basic(appName, rName1, rName2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "configuration_profile_ids.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "configuration_profile_ids.*", "aws_appconfig_configuration_profile.test_1", "configuration_profile_id"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "configuration_profile_ids.*", "aws_appconfig_configuration_profile.test_2", "configuration_profile_id"),
				),
			},
		},
	})
}

func testAccConfigurationProfilesDataSourceConfig_basic(appName, rName1, rName2 string) string {
	return acctest.ConfigCompose(
		testAccApplicationConfig_name(appName),
		fmt.Sprintf(`
resource "aws_appconfig_configuration_profile" "test_1" {
  application_id = aws_appconfig_application.test.id
  name           = %[1]q
  location_uri   = "hosted"
}

resource "aws_appconfig_configuration_profile" "test_2" {
  application_id = aws_appconfig_application.test.id
  name           = %[2]q
  location_uri   = "hosted"
}

data "aws_appconfig_configuration_profiles" "test" {
  application_id = aws_appconfig_application.test.id
  depends_on     = [aws_appconfig_configuration_profile.test_1, aws_appconfig_configuration_profile.test_2]
}
`, rName1, rName2))
}
