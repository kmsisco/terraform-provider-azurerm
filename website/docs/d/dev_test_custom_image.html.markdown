---
layout: "azurerm"
page_title: "Azure Resource Manager: azurerm_dev_test_custom_image"
sidebar_current: "docs-azurerm-datasource-dev-test-custom-image"
description: |-
  Gets information about an existing Dev Test Lab Custom Image.
---

# Data Source: azurerm_dev_test_custom_image

Use this data source to access information about an existing Dev Test Lab Custom Image.

## Example Usage

```hcl
data "azurerm_dev_test_custom_image" "test" {
  name                = "customimage"
  lab_name            = "examplelab"
  resource_group_name = "example-resource"
}

output "custom_image_id" {
  value = "${data.azurerm_dev_test_custom_image.test.id}
}
```

## Argument Reference

* `name` - (Required) Specifies the name of the Custom Image.
* `lab_name` - (Required) Specifies the name of the Dev Test Lab.
* `resource_group_name` - (Required) Specifies the name of the resource group that contains the Custom Image.
