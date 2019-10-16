---
layout: "azurerm"
page_title: "Azure Resource Manager: azurerm_dev_test_custom_image"
sidebar_current: "docs-azurerm-resource-dev-test-custom-image"
description: |-
  Manages a Dev Test Lab Custom Image.
---

# azurerm_dev_test_custom_image

Manages a Dev Test Lab Custom Image.

## Example Usage

```hcl
resource "azurerm_resource_group" "test" {
  name     = "example-resources"
  location = "West US"
}

resource "azurerm_dev_test_lab" "test" {
  name                = "examplelab"
  location            = "${azurerm_resource_group.test.location}"
  resource_group_name = "${azurerm_resource_group.test.name}"
}

resource "azurerm_dev_test_virtual_network" "test" {
  name                = "examplevnet"
  lab_name            = "${azurerm_dev_test_lab.test.name}"
  resource_group_name = "${azurerm_resource_group.test.name}"

  subnet {
    use_public_ip_address           = "Allow"
    use_in_virtual_machine_creation = "Allow"
  }
}

resource "azurerm_dev_test_linux_virtual_machine" "test" {
  name                   = "examplevm"
  lab_name               = "${azurerm_dev_test_lab.test.name}"
  resource_group_name    = "${azurerm_resource_group.test.name}"
  location               = "${azurerm_resource_group.test.location}"
  size                   = "Standard_B1ms"
  username               = "user"
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
  name                = "customimage"
  lab_name            = "${azurerm_dev_test_lab.test.name}"
  resource_group_name = "${azurerm_resource_group.test.name}"
  location            = "${azurerm_resource_group.test.location}"
  author              = "Author"
  description         = "A test custom image"

  vm {
    source_vm_id  = "${azurerm_dev_test_linux_virtual_machine.test.id}"
    linux_os_info {
      linux_os_state = "DeprovisionRequested"
    }
  }
}
```

## Argument Reference

* `name` - (Required) Specifies the name of the Custom Image.
* `lab_name` - (Required) Specifies the name of the Dev Test Lab.
* `resource_group_name` - (Required) Specifies the name of the resource group that contains the Custom Image.
* `location` - (Required) The location of the custom image.
* `author` - The name of the creator of the custom image.
* `description` - A text description for the custom image.
* `vm` - A `vm` block as defined below.

---

A `vm` block supports the following:

* `source_vm_id` - (Required) The ID of the virtual machine used to create the custom image.
* `linux_os_info` - A `linux_os_info` block as defined below.
* `windows_os_info` - A `windows_os_info` block as defined below.

---

A `linux_os_info` block supports the following:

* `linux_os_state` - The deprovision status of the Linux OS. Possible values are `DeprovisionApplied`, `DeprovisionRequested` or `NonDeprovisioned`.

---

A `windows_os_info` block supports the following:

* `windows_os_state` - The sysprep status of the Windows OS. Possible values are `SysprepApplied`, `SysprepRequested` or `NonSysprepped`.

## Attributes Reference

* `id` - The ID of the Custom Image.

## Import

Dev Test Custom Images can be imported using the `resource id`, e.g.

```shell
terraform import azurerm_dev_test_custom_image.customimage1 /subscriptions/00000000-0000-0000-0000-000000000000/resourcegroups/example-resources/providers/microsoft.devtestlab/labs/example-dtl/customimages/customimage
```
