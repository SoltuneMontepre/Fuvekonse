variable "project_name" {
  description = "The name of the project"
  type        = string
}

variable "cors_allowed_origins" {
  description = "List of allowed origins for CORS"
  type        = list(string)
  default     = []
}

variable "general_service_function_name" {
  description = "Name of the general service Lambda function"
  type        = string
}

variable "general_service_invoke_arn" {
  description = "Invoke ARN of the general service Lambda function"
  type        = string
}

variable "rbac_service_function_name" {
  description = "Name of the RBAC service Lambda function"
  type        = string
}

variable "rbac_service_invoke_arn" {
  description = "Invoke ARN of the RBAC service Lambda function"
  type        = string
}
