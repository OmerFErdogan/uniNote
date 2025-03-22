# 📌 techContext.md

## Overview
UniNotes' technical infrastructure is built on a **Go** based backend and a **VPS-focused hosting** model. The goal is to establish a structure suitable for medium-scale traffic that is performant and developer-friendly.

## Programming Language and Architecture
- **Language:** Go (1.18+) is suitable for creating high-performance web services with strong concurrency features and lightweight threads (goroutines).  
- **Architectural Principle:** Clean Architecture; clearly separates domain and infrastructure layers, prioritizing testability and sustainability.  

## Frameworks and Layers
- **Web Framework:** A lightweight and performant HTTP framework like Go Fiber or Gin.  
- **Use Case Layer:** Organizes application-specific scenarios independently from database or file system details.  
- **Adapter Layer:** Contains implementations of external components such as PostgreSQL, file system, cache.  

## Database
- **Choice:** PostgreSQL (>= version 13) is recommended.  
- **VPS Installation:** PostgreSQL is installed and managed on the VPS with the help of documentation and automation (e.g., Ansible).  
- **Size Management:** A single database is sufficient at the MVP stage. Vertical scaling (higher CPU/RAM) or adding read replicas can be considered when traffic and data volume increase.

## File Storage
- **Local File System:** PDFs and images are kept in a directory like `/data/` or similar on the VPS.  
- **Backup & Access:** Disk capacity and I/O performance are considered for large file uploads. Regular backups (e.g., rsync, duplicity) are recommended.  
- **Expansion:** A more extensive object storage (MinIO, Ceph, etc.) or cloud-based S3 variant can be integrated in the future.

## Authentication
- **JWT Based:** Each API request performs authentication through a token in the HTTP header.  
- **Encryption:** User passwords are stored hashed with an algorithm like `bcrypt`. Salt and iteration numbers are securely set.

## VPS Configuration
- **Server Selection:** A VPS plan with 2-4 CPUs and 4-8 GB RAM from providers like DigitalOcean, Hetzner, Linode, or similar is suitable for the start, depending on traffic expectations.  
- **Deployment:**  
  - Running with Docker or direct Go binary  
  - Automated testing + SSH deployment via CI/CD pipeline (GitHub Actions, GitLab CI)  
- **Security:**  
  - SSH key-based access, limited port opening (80/443/22)  
  - HTTPS with Let's Encrypt certificates  
  - Basic security measures with tools like iptables/ufw  

## Collaboration and Real-Time Features
- **Technology Choices:** WebSocket, SSE, or Socket.IO Go library  
- **Concurrency Management:** Use cases for one goroutine per connection, using channels or mutexes to protect shared data structures  
- **Versioning:** Conflict resolution strategies when multiple users edit the same note (Operational Transform, CRDT, etc.)

## Logging & Monitoring
- **Log Collection:** Logging in JSON format to standard output (stdout) or file. External log systems (ELK, Loki) can be integrated if needed.  
- **Monitoring:** CPU, memory, request count, error rates are examined by collecting metrics with tools like Prometheus and Grafana.  
- **Alert Mechanisms:** Alerts can be sent via Slack, email, or third-party notification service in cases of high error rates, abnormal resource usage, etc.

## Constraints and Risks
- **VPS Limits:**  
  - Disk capacity and network bandwidth may be more limited compared to cloud storage like S3.  
  - No automatic scaling; manual intervention or additional VPS is required.  
- **Multiple VPS:** When horizontal growth is needed, load balancer, DNS round-robin, or reverse proxy configuration (NGINX, HAProxy, Caddy, etc.) comes into play.  
- **Additional Maintenance Load:** Tasks such as patch management, security patches, data backup are the responsibility of the project team.

## Future Plans
1. **Microservice Tendency:** Scaling components like user, notes, PDF processing as separate services.  
2. **Container Orchestration:** Gaining advantages like automatic pod scaling, self-healing by transitioning to Kubernetes.  
3. **Advanced Cache / Queue:** Managing scenarios under heavy load with tools like Redis, RabbitMQ, or NATS.  
4. **Additional 3rd Party Services:** Email sending, messaging, push notification, mobile application integrations.