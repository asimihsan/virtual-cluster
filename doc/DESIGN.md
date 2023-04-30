# DESIGN.md

## Problem Statement
The current process of managing the startup of containerized microservices and their dependencies is fragmented and lacks a centralized solution. This leads to difficulties in initializing services, running end-to-end integration tests, and handling service failures during startup or health check processes.

## High-Level Solution
Create a new project called `virtual-cluster` to manage the startup, initialization, and testing of containerized microservices and their dependencies, such as LocalStack and PostgreSQL. This repository will utilize custom Domain Specific Languages (DSLs) for service configuration and provide a seamless developer experience by collecting logs and OTEL traces for efficient debugging.

## Solution Details

### Repository Structure

The `virtual-cluster` project is the command-line tool.

Customers will create a config repo e.g. `virtual-cluster-config` that contains the following:

- `.services`: A file listing the services, their repositories, branch/tag/commit, and directory within the repo (or root). This file will utilize a custom ANTLR-based DSL for consistency and ease of use.

Customers will create the following files in each of their service repositories:

- `.vcluster`: A file included in each service's repository containing information such as service dependencies, health check endpoints, startup sequences, and other necessary details. This file will also use a custom ANTLR-based DSL.

### Custom DSLs
The custom DSLs for the `.services` and `.vcluster` files will be created using the ANTLR (ANother Tool for Language Recognition) framework. This will ensure clear semantics and excellent static analysis capabilities. The DSL will be similar to HCL but have a strict lexical grammar to make static analysis and IDE auto-suggestion easier. The DSL must be easy to use, learn, and train for.

### Log and Trace Collection
To provide a seamless developer experience, `virtual-cluster` will collect logs and OTEL traces for efficient debugging. Instead of forwarding logs, the repository will run a local service responsible for collecting and analyzing logs and traces. This will ensure easy access to relevant information for developers.

### Service Interaction
In order to execute the `virtual-cluster`, Git submodules or cloning of the Git repositories for each service will be required. This will enable the repository to manage and interact with the individual services effectively.
