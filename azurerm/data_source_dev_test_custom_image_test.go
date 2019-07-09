package azurerm

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
)

func TestAccDataSourceAzureRMDevTestCustomImage_basic(t *testing.T) {
	dataSourceName := "data.azurerm_dev_test_custom_image.test"
	rInt := tf.AccRandTimeInt()
	location := testLocation()

	name := fmt.Sprintf("acctestimage%d", rInt)
	labName := fmt.Sprintf("acctestdtl%d", rInt)
	resGroup := fmt.Sprintf("acctestrg-%d", rInt)
	imageID := fmt.Sprintf("/subscriptions/%s/resourcegroups/%s/providers/microsoft.devtestlab/labs/%s/customimages/%s", os.Getenv("ARM_SUBSCRIPTION_ID"), resGroup, labName, name)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceDevTestCustomImage_basic(rInt, location),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "name", name),
					resource.TestCheckResourceAttr(dataSourceName, "lab_name", labName),
					resource.TestCheckResourceAttr(dataSourceName, "resource_group_name", resGroup),
					resource.TestCheckResourceAttr(dataSourceName, "author", "acctest"),
					resource.TestCheckResourceAttr(dataSourceName, "description", "A test custom image"),
					resource.TestCheckResourceAttr(dataSourceName, "id", imageID),
				),
			},
		},
	})
}

func testAccDataSourceDevTestCustomImage_basic(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurerm_resource_group" "test" {
  name     = "acctestrg-%d"
  location = "%s"
}

resource "azurerm_dev_test_lab" "test" {
  name                = "acctestdtl%d"
  location            = "${azurerm_resource_group.test.location}"
  resource_group_name = "${azurerm_resource_group.test.name}"
}

resource "azurerm_dev_test_virtual_network" "test" {
  name                = "acctestvn%d"
  lab_name            = "${azurerm_dev_test_lab.test.name}"
  resource_group_name = "${azurerm_resource_group.test.name}"

  subnet {
    use_public_ip_address           = "Allow"
    use_in_virtual_machine_creation = "Allow"
  }
}

resource "azurerm_dev_test_linux_virtual_machine" "test" {
  name                   = "acctestvm-vm%d"
  lab_name               = "${azurerm_dev_test_lab.test.name}"
  resource_group_name    = "${azurerm_resource_group.test.name}"
  location               = "${azurerm_resource_group.test.location}"
  size                   = "Standard_B1ms"
  username               = "acct5stU5er"
  password               = "Pa$$w0rd1234!"
  lab_virtual_network_id = "${azurerm_dev_test_virtual_network.test.id}"
  lab_subnet_name        = "${azurerm_dev_test_virtual_network.test.subnet.0.name}"
  storage_type           = "Premium"

  gallery_image_reference {
    offer     = "UbuntuServer"
    publisher = "Canonical"
    sku       = "18.04-LTS"
    version   = "latest"
  }
}

resource "azurerm_dev_test_custom_image" "test" {
  name                = "acctestimage%d"
  lab_name            = "${azurerm_dev_test_lab.test.name}"
  resource_group_name = "${azurerm_resource_group.test.name}"
  location            = "${azurerm_resource_group.test.location}"
  author              = "acctest"
  description         = "A test custom image"

  vm {
    source_vm_id  = "${azurerm_dev_test_linux_virtual_machine.test.id}"
    linux_os_info {
      linux_os_state = "DeprovisionRequested"
    }
  }
}

data "azurerm_dev_test_custom_image" "test" {
  name                = "${azurerm_dev_test_custom_image.test.name}"
  lab_name            = "${azurerm_dev_test_lab.test.name}"
  resource_group_name = "${azurerm_resource_group.test.name}"
}
`, rInt, location, rInt, rInt, rInt, rInt)
}
