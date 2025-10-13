# K8s Reporter

## Overview

K8s Reporter is a comprehensive monitoring and visualization platform for Kubernetes clusters. The project aims to provide an intuitive interface for teams to understand and monitor their Kubernetes infrastructure, making cluster information accessible to both technical and non-technical stakeholders.

## What is K8s Reporter?

K8s Reporter helps organizations gain visibility into their Kubernetes environments by collecting, organizing, and presenting cluster information in a user-friendly way. Whether you need to check the health of your deployments, monitor resource usage, or understand the state of your infrastructure, K8s Reporter provides the tools to do so efficiently.

## Architecture

The project is built on a modern, scalable architecture consisting of three main components:

1. **Cluster Data Collection Service** - A Go-based service that connects to Kubernetes clusters and collects real-time information about pods, deployments, services, and other resources.

2. **GraphQL API** - A flexible API layer that processes and serves cluster data, enabling efficient queries and real-time updates.

3. **Web Frontend** - An intuitive web interface that visualizes cluster information through dashboards, charts, and interactive components.

4. **Authentication & Authorization** - Security layer ensuring that only authorized users can access cluster information, with role-based access control for different permission levels.

## Current Status

The project is in active development. Currently, the **Cluster Data Collection Service** (k8s-go-reporter) has been implemented and is capable of gathering information from Kubernetes clusters.

To testing pipeline in localhost use act, case whant test projeto, use make.

**In Progress:**
- GraphQL API development
- Frontend web application
- Authentication and authorization system

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
