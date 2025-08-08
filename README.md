# Get Resource CLI

## Overview

Get Resource CLI is a Golang TUI to fetch terraform provider resource schema and transform.

The CLI provides a TUI (https://github.com/charmbracelet/bubbletea) with fuzzy autocomplete to select a provider and resource in the active terraform project.

Once selected, the TUI displays every argument and attribute (including nested attributes) of the selection and offers several built-in transformers to work with the resource schema.

## User interface and workflows

### Main command

Arguments:
- path (Default PWD): A path to the terraform configuration to work with.

Example workflow from main:

1. The cli is started with a path (pwd by default) and verifies if a Terraform configuration is available. (exit 1 if no terraform configuration found in path)
2. The cli will detect which flavor (Terraform or opentofu - default to tofu) is available and ensure `init` is ran (os.exec) to download all the providers used in the configuration.
4. behind the scenes, the cli loads the providers schema (either from a cache on disk ~ .resource-cache/ by default) or by invoking the terraform provider schema command (https://developer.hashicorp.com/terraform/cli/commands/providers/schema), writing the json output to cache and deserializing it into memory.
3. The user is presented with a list of available providers in the configuration and can either navigate to a listed provider (or start typing and the provider list is filtered down (fuzzy match) to providers matching what the user typed). Press enter to select
4. Once a provider is selected, the User is presented with a simple view with 4 sections: Data Sources, Ephemeral Resources, Provider Functions and Resources. For each section, the same User experience (either use navigation keys to select a provider or fuzzy match type to filter down the list), allows the user to select a resource within the provider (consider the TF)
5. Once a resource is selected, the user is presented with 2 sections: Arguments and Attributes, navigating between them shows a list view next to it which the user can switch to and back to the section navigator.
6. When in a specific resource section, the user can trigger a built-in "transformation"
7. The first (and only) built-in transformation is "convert to HCL".

Note on TUI representation of provider schemas:

The provider schemas are converted to string representations of the nested types which fit on screen. At the moment this is a very simple conversion to string which we may iterate on in the future.

### Sub command: init

This command allows a user to init a directory (bail if not empty) by listing popular terraform providers (aws, gcp, azure to start with) and automatically generating a `providers.tf` file defining the Terraform configuration block (required providers, ... )

## Transformations

### Convert to HCL

The convert to HCL transformation takes in a Resource Section (Attributes or Arguments).

####  When passed in an Arguments Resource Section:

The transformation will generate "terraform variable" HCL blocks for every element.

Example output
```hcl
variable "foo" {
    type = <attribute type>
    description = <attribute description>
}
```

If the attribute is a nested type, we will create an equivalent terraform Input variable "object". Refer to https://developer.hashicorp.com/terraform/language/values/variables#declaring-an-input-variable

#### When passed in an Attributes Resource Section:


The transformation will generate "terraform output" HCL blocks for every element.

Example output
```hcl
# optional nested schema TUI representation
output "foo" {
    value = <attribute ref> # for example aws_instance.server.private_ip
}
```

If the attribute is a nested type, we just create one output for the root attribute.

We also include the nested schema in the same way it is represented in the TUI as a comment section above the output for end user reference.

Refer to https://developer.hashicorp.com/terraform/language/values/outputs
