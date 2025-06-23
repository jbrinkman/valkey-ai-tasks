# Example Agent Prompts for Valkey MCP Task Management

This document provides example prompts for AI agents to effectively use the Valkey MCP Task Management system, with a focus on utilizing the notes functionality.

## Basic Project and Task Management

### Creating a Plan with Initial Notes

```
Create a new plan called "E-commerce Platform" for application "retail-app" with the following notes:

# E-commerce Platform Project

## Objectives
- Build a scalable e-commerce platform
- Support multiple payment gateways
- Implement inventory management
- Create a responsive user interface

## Technical Requirements
- Use microservices architecture
- Implement CI/CD pipeline
- Ensure GDPR compliance
- Support mobile and desktop views

## Timeline
- Phase 1: Core functionality (4 weeks)
- Phase 2: Payment integration (2 weeks)
- Phase 3: Inventory management (3 weeks)
- Phase 4: UI/UX improvements (3 weeks)
```

### Creating Tasks with Notes

```
For the "E-commerce Platform" plan, create the following tasks with appropriate priorities:

1. "Set up microservices architecture" (high priority) with notes:
   # Microservices Setup

   ## Components
   - User service
   - Product service
   - Order service
   - Payment service
   - Inventory service

   ## Technology Stack
   - Go for backend services
   - PostgreSQL for persistent data
   - Redis for caching
   - gRPC for inter-service communication

   ## Implementation Steps
   1. Define service boundaries
   2. Create service templates
   3. Set up Docker containers
   4. Configure service discovery

2. "Implement product catalog" (medium priority) with notes:
   # Product Catalog Implementation

   ## Data Model
   ```json
   {
     "product": {
       "id": "string",
       "name": "string",
       "description": "string",
       "price": "number",
       "categories": ["string"],
       "attributes": {
         "key": "value"
       }
     }
   }
   ```

   ## API Endpoints
   - GET /products - List all products
   - GET /products/{id} - Get product details
   - POST /products - Create new product
   - PUT /products/{id} - Update product
   - DELETE /products/{id} - Delete product

3. "Design user authentication flow" (high priority)
```

## Maintaining Context Between Sessions

### Updating Plan Notes with Progress

```
Update the notes for the "E-commerce Platform" plan to include the following information:

# E-commerce Platform Project

## Objectives
- Build a scalable e-commerce platform
- Support multiple payment gateways
- Implement inventory management
- Create a responsive user interface

## Technical Requirements
- Use microservices architecture
- Implement CI/CD pipeline
- Ensure GDPR compliance
- Support mobile and desktop views

## Timeline
- Phase 1: Core functionality (4 weeks) - IN PROGRESS
- Phase 2: Payment integration (2 weeks)
- Phase 3: Inventory management (3 weeks)
- Phase 4: UI/UX improvements (3 weeks)

## Progress Update (2025-06-23)
- Microservices architecture design completed
- Database schema defined for user and product services
- Docker containers set up for development environment
- CI pipeline configured with GitHub Actions

## Challenges
- Need to resolve service discovery issues in Docker environment
- Evaluating payment gateway options (Stripe vs PayPal)
```

### Documenting Technical Decisions in Task Notes

```
Update the notes for the "Set up microservices architecture" task with the following technical decisions:

# Microservices Setup

## Components
- User service
- Product service
- Order service
- Payment service
- Inventory service

## Technology Stack
- Go for backend services
- PostgreSQL for persistent data
- Redis for caching
- gRPC for inter-service communication

## Implementation Steps
1. Define service boundaries ✅
2. Create service templates ✅
3. Set up Docker containers ✅
4. Configure service discovery 🔄

## Technical Decisions

### Service Discovery
After evaluating Consul, etcd, and Kubernetes native service discovery, we've decided to use Kubernetes service discovery for the following reasons:
- Simplifies deployment model
- Native integration with our planned infrastructure
- Reduces operational complexity
- Built-in load balancing

### Database Per Service
Each microservice will have its own database to ensure:
- Loose coupling between services
- Independent scaling
- Isolation of data concerns

### API Gateway
We'll implement an API Gateway using Envoy because:
- High performance
- Extensible architecture
- Strong community support
- Built-in observability features

## Code Examples

### Service Template

```go
package main

import (
    "log"
    "net"
    "os"

    "google.golang.org/grpc"
    pb "github.com/example/service/proto"
)

func main() {
    port := os.Getenv("PORT")
    if port == "" {
        port = "50051"
    }
    
    lis, err := net.Listen("tcp", ":"+port)
    if err != nil {
        log.Fatalf("failed to listen: %v", err)
    }
    
    s := grpc.NewServer()
    pb.RegisterServiceServer(s, &server{})
    
    log.Printf("server listening at %v", lis.Addr())
    if err := s.Serve(lis); err != nil {
        log.Fatalf("failed to serve: %v", err)
    }
}
```
```

## Collaborative Workflows

### Handoff Between Team Members

```
Review the "E-commerce Platform" plan and all its tasks. Update the plan notes to include this handoff information:

# Team Handoff (2025-06-23)

## Current Status
- Microservices architecture design is complete
- User authentication service is 70% implemented
- Product catalog API is functional but needs pagination
- CI/CD pipeline is configured but needs test automation

## Immediate Next Steps
1. Complete user authentication service
2. Add pagination to product catalog API
3. Implement test automation in CI/CD pipeline
4. Start work on payment service integration

## Known Issues
- Performance bottleneck in product search functionality
- Docker container for Redis occasionally fails to start
- Authentication token expiration needs to be configurable

## Contact Information
- Backend questions: alex@example.com
- DevOps questions: sam@example.com
- Project management: jamie@example.com
```

### Tracking Requirements Changes

```
Update the notes for the "Implement product catalog" task to reflect new requirements:

# Product Catalog Implementation

## Data Model
```json
{
  "product": {
    "id": "string",
    "name": "string",
    "description": "string",
    "price": "number",
    "sale_price": "number",  // NEW: Added for promotional pricing
    "categories": ["string"],
    "attributes": {
      "key": "value"
    },
    "inventory_status": "string",  // NEW: in_stock, low_stock, out_of_stock
    "images": ["string"],  // NEW: Array of image URLs
    "rating": "number",  // NEW: Average customer rating
    "reviews_count": "number"  // NEW: Number of customer reviews
  }
}
```

## API Endpoints
- GET /products - List all products
- GET /products/{id} - Get product details
- POST /products - Create new product
- PUT /products/{id} - Update product
- DELETE /products/{id} - Delete product
- GET /products/category/{category} - NEW: Filter products by category
- GET /products/search - NEW: Search products with filters

## Requirements Changes (2025-06-23)
1. Added support for promotional pricing (sale_price field)
2. Added inventory status tracking
3. Added support for multiple product images
4. Added customer ratings and reviews count
5. Added new API endpoints for category filtering and advanced search

## Implementation Impact
- Database schema needs to be updated
- Additional API endpoints need to be implemented
- UI components need to be updated to display new fields
- Search functionality needs to be enhanced
```

## Debugging and Troubleshooting

### Documenting Issues and Solutions

```
Update the notes for the "Set up microservices architecture" task to include troubleshooting information:

# Troubleshooting Log

## Issue: Service Discovery Failure
**Date:** 2025-06-22
**Symptoms:** Services unable to discover and communicate with each other
**Root Cause:** DNS resolution issues in Kubernetes cluster
**Solution:**
1. Updated CoreDNS configuration
2. Added explicit DNS policy to pod specifications
3. Implemented retry logic with exponential backoff

**Code Fix:**
```go
func getServiceClient(ctx context.Context, serviceName string) (client.ServiceClient, error) {
    var serviceClient client.ServiceClient
    var err error
    
    backoff := retry.NewExponentialBackOff()
    backoff.MaxElapsedTime = 30 * time.Second
    
    retryFn := func() error {
        serviceClient, err = client.NewServiceClient(ctx, serviceName)
        return err
    }
    
    err = retry.Retry(retryFn, backoff)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to service %s after retries: %w", serviceName, err)
    }
    
    return serviceClient, nil
}
```

## Issue: Database Connection Pooling
**Date:** 2025-06-23
**Symptoms:** Services experiencing intermittent database timeouts under load
**Root Cause:** Insufficient connection pool configuration
**Solution:**
1. Increased max connections in pool
2. Added connection lifetime management
3. Implemented circuit breaker pattern

**Configuration:**
```yaml
database:
  max_open_conns: 25
  max_idle_conns: 10
  conn_max_lifetime: 5m
  conn_max_idle_time: 1m
```
```

## Patterns for Effective Notes Usage by Agents

### Project Planning Pattern

```
Create a new plan called "Mobile App Redesign" for application "mobile-client" with the following structured notes:

# Mobile App Redesign Project

## Project Overview
[Brief description of the plan goals and scope]

## Key Stakeholders
- Product Owner: [Name]
- Designer: [Name]
- Lead Developer: [Name]
- QA Lead: [Name]

## Requirements
[List of key requirements]

## Design Documents
[Links to design documents, mockups, etc.]

## Technical Approach
[Overview of the technical approach]

## Timeline
[Project timeline with key milestones]

## Status Updates
[Section for regular status updates]

## Decisions Log
[Section for documenting key decisions]

## Open Questions
[Section for tracking open questions]
```

### Task Implementation Pattern

```
Create a task called "Implement User Authentication" for the "Mobile App Redesign" plan with high priority and the following structured notes:

# User Authentication Implementation

## Requirements
[List of specific requirements for this task]

## Technical Approach
[Description of the technical approach]

## Dependencies
[List of dependencies]

## Implementation Steps
[Step-by-step implementation plan]

## Testing Strategy
[Description of how this will be tested]

## Code Snippets
[Section for relevant code snippets]

## Resources
[Links to relevant documentation, examples, etc.]

## Progress Updates
[Section for tracking progress]

## Review Notes
[Section for review feedback]
```

### Bug Tracking Pattern

```
Create a task called "Fix Login Screen Crash" for the "Mobile App Redesign" plan with high priority and the following structured notes:

# Login Screen Crash Bug

## Bug Description
[Detailed description of the bug]

## Steps to Reproduce
1. [Step 1]
2. [Step 2]
3. [Step 3]

## Environment
- Device: [Device information]
- OS Version: [OS version]
- App Version: [App version]

## Stack Trace
```
[Stack trace goes here]
```

## Root Cause Analysis
[Analysis of the root cause]

## Fix Implementation
[Description of the fix]

## Verification Steps
[Steps to verify the fix]

## Regression Testing
[Plan for regression testing]
```
