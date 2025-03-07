package kafkaconnect_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/kafkaconnect"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccKafkaConnectCustomPluginDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mskconnect_custom_plugin.test"
	dataSourceName := "data.aws_mskconnect_custom_plugin.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, kafkaconnect.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, kafkaconnect.EndpointsID),
		CheckDestroy:             nil,
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomPluginDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "description", dataSourceName, "description"),
					resource.TestCheckResourceAttrPair(resourceName, "latest_revision", dataSourceName, "latest_revision"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "state", dataSourceName, "state"),
				),
			},
		},
	})
}

func testAccCustomPluginDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccCustomPluginBaseConfig(rName, false), fmt.Sprintf(`
resource "aws_mskconnect_custom_plugin" "test" {
  name         = %[1]q
  content_type = "ZIP"

  location {
    s3 {
      bucket_arn = aws_s3_bucket.test.arn
      file_key   = aws_s3_object.test.key
    }
  }
}

data "aws_mskconnect_custom_plugin" "test" {
  name = aws_mskconnect_custom_plugin.test.name
}
`, rName))
}
