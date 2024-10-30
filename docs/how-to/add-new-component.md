# Add New Component

This guide explains how to add a new infrastructure component to the project.

## Component Types

You can add these types of components:
1. New Terraform module
2. New application configuration
3. New environment configuration

## Adding a New Module

1. Create module structure:

```bash
mkdir modules/new-module
cd modules/new-module
touch {main,variables,outputs}.tf
```

2. Define module interface:

```hcl
# variables.tf
variable "environment" {
  description = "Environment name"
  type        = string
}

variable "project_name" {
  description = "Project name"
  type        = string
}

# Add other required variables
```

3. Implement module logic in main.tf
4. Define outputs in outputs.tf
5. Add documentation:
   - README.md with usage instructions
   - Update architecture.md
   - Add variables to variables.md

## Testing New Components

1. Add unit tests:

```go
package test

func TestNewModule(t *testing.T) {
    t.Parallel()
    // Add test implementation
}
```

2. Add the module to e2e tests if needed

3. Run tests:
```bash
cd test && go test -v ./...
```

## Integration

1. Add to complete example:

```hcl
module "new_module" {
  source = "../../modules/new-module"
  
  environment  = var.environment
  project_name = var.project_name
  # Add other variables
}
```

2. Update CI/CD pipeline if needed
3. Update documentation
4. Create example usage

## Best Practices

1. Follow existing patterns
2. Include proper tagging
3. Add security controls
4. Enable monitoring
5. Document all variables
