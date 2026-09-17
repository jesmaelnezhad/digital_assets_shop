# 01 — RED Server Provisioning

**Date:** 2026-09-04  
**Server:** RED (194.5.206.106)  
**Role:** Serving cluster for Pawradise

## Server Specification

- OS: Ubuntu 24.04.4 LTS
- CPU: 2 cores
- RAM: 1.9 GB
- Disk: TBD
- Access: SSH root@194.5.206.106 (key-based, from BLUE)
- k3s: not installed
- Docker: not installed
- containerd: not installed

## Initial State

Fresh server, no packages installed beyond base Ubuntu. No k3s, no Docker, no containerd.

## Purpose

This server will host:
- k3s cluster (single node, no traefik, no local-path-provisioner)
- PostgreSQL 16 (StatefulSet in k3s)
- Host nginx on port 80 (path-based routing to k3s NodePorts)
- Backend pods (staging + production)
- Frontend pods (staging + production)

BLUE will handle all builds, image exports, and test execution.
