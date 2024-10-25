#!/usr/bin/env bash

# Set strict error handling
set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Logging functions
log_error() {
    echo -e "${RED}ERROR: $1${NC}" >&2
}

log_success() {
    echo -e "${GREEN}SUCCESS: $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}WARNING: $1${NC}"
}

log_info() {
    echo -e "INFO: $1"
}

# Check if a command exists
check_command() {
    if ! command -v "$1" &> /dev/null; then
        log_error "$1 could not be found"
        return 1
    fi
    log_success "$1 is installed"
    return 0
}

# Check Python version
check_python() {
    if ! check_command "python3"; then
        log_error "Python3 is required but not installed"
        exit 1
    fi

    local python_version
    python_version=$(python3 --version | cut -d' ' -f2)
    log_success "Python version ${python_version} is installed"
}

# Setup virtual environment
setup_venv() {
    local venv_dir="venv"
    
    if [ ! -d "$venv_dir" ]; then
        log_info "Creating virtual environment..."
        python3 -m venv "$venv_dir"
    fi

    log_info "Activating virtual environment..."
    # shellcheck source=/dev/null
    source "${venv_dir}/bin/activate"

    log_info "Installing/Updating AWS CLI..."
    pip install --upgrade awscli
    
    log_success "Virtual environment is setup and AWS CLI is installed/updated"
}

# Check Terraform version
check_terraform() {
    if ! check_command "terraform"; then
        log_error "Terraform is required but not installed"
        log_info "Please install Terraform ${REQUIRED_TF_VERSION} or higher"
        log_info "Visit: https://developer.hashicorp.com/terraform/downloads"
        exit 1
    fi

    local tf_version
    tf_version=$(terraform version -json | jq -r '.terraform_version')

    log_success "Terraform version ${tf_version} is installed"
}

# Check Go installation
check_go() {
    if ! check_command "go"; then
        log_error "Go is required but not installed"
        exit 1
    fi

    local go_version
    go_version=$(go version | cut -d' ' -f3)
    log_success "Go ${go_version} is installed"
}

# Check AWS credentials
check_aws_credentials() {
    log_info "Checking AWS credentials..."
    
    if ! aws sts get-caller-identity &> /dev/null; then
        log_error "AWS credentials are not properly configured"
        log_error "Please configure AWS credentials using 'aws configure' or set appropriate environment variables"
        exit 1
    fi
    
    local account_id
    account_id=$(aws sts get-caller-identity --query "Account" --output text)
    log_success "AWS credentials are properly configured for account ${account_id}"
}

# Initialize and check backend infrastructure
setup_backend() {
    log_info "Checking backend configuration..."

    # Check if backend tfvars exists
    if [ ! -f "backend/terraform.tfvars" ]; then
        log_error "backend/terraform.tfvars not found"
        exit 1
    fi

    # Extract bucket name from tfvars
    local bucket_name
    bucket_name=$(grep 'state_bucket_name' backend/terraform.tfvars | cut -d'=' -f2 | tr -d ' "')

    if [ -z "$bucket_name" ]; then
        log_error "Could not find state_bucket_name in backend/terraform.tfvars"
        exit 1
    fi

    # Check if backend state exists
    if aws s3api head-bucket --bucket "$bucket_name" 2>/dev/null; then
        log_success "Backend infrastructure exists"

        # Verify the state is accessible
        (
            cd backend
            if ! terraform init; then
                log_error "Failed to initialize backend configuration"
                exit 1
            fi

            if ! terraform plan -detailed-exitcode &>/dev/null; then
                log_warning "Backend infrastructure exists but might need updates"
                read -rp "Would you like to apply any pending changes? (y/n) " update_backend
                if [[ $update_backend =~ ^[Yy]$ ]]; then
                    if ! terraform apply; then
                        log_error "Failed to update backend infrastructure"
                        exit 1
                    fi
                    log_success "Backend infrastructure updated successfully"
                fi
            else
                log_success "Backend infrastructure is up to date"
            fi
        )
        return 0
    fi

    log_warning "Backend infrastructure does not exist"
    read -rp "Would you like to create the backend infrastructure? (y/n) " create_backend

    if [[ $create_backend =~ ^[Yy]$ ]]; then
        log_info "Initializing backend infrastructure..."

        # Initialize backend directory
        (
            cd backend
            if ! terraform init; then
                log_error "Failed to initialize backend configuration"
                exit 1
            fi

            if ! terraform apply; then
                log_error "Failed to create backend infrastructure"
                exit 1
            fi
        )

        log_success "Backend infrastructure created successfully"
    else
        log_error "Cannot proceed without backend infrastructure"
        exit 1
    fi
}

# Main execution
main() {
    log_info "Starting bootstrap process..."
    
    # Check tooling dependencies
    check_python
    setup_venv
    check_terraform
    check_go

    # Setup AWS credentials and backend infrastructure
    check_aws_credentials
    setup_backend
    
    log_success "Bootstrap completed successfully!"
    log_info "You can now proceed with terraform init and other commands"
}

main "$@"
