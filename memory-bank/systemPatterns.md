# 📌 systemPatterns.md

## Architectural Strategy
UniNotes' server-side (backend) will be developed in Go, aiming for high performance and an easily maintainable structure. **Clean Architecture** principles will be adopted. This ensures that the business logic (domain) in the code base remains independent of infrastructure details, making it easier to test and sustainable in the long term.

### Goals
1. **Independence:** Minimal dependency on frameworks and external dependencies; the domain (business rules) layer is isolated from external factors.  
2. **Testability:** Since communication occurs through interfaces, business rules can be easily tested without real databases or services.  
3. **Ease of Maintenance:** The layered structure provides readability and modularity even as the project grows.  
4. **Adaptability:** When making new feature or technology changes, the core domain code is minimally affected.

---

## Layers & Directory Structure

We suggest **4 basic layers** by structuring the Clean Architecture approach through Go packages:

1. **Domain (Core) Layer**  
   - **What it includes?**  
     - Entity (model) definitions, business rules (domain services), data types.  
   - **Dependencies:**  
     - Not directly dependent on any third-party library or infrastructure. Only utilizes Go's standard library.  
   - **Responsibility:**  
     - Covers the most critical business logic of the application; rules for notes, PDFs, user accounts are defined in this layer.

2. **Use Case (Application) Layer**  
   - **What it includes?**  
     - Services that manage specific scenarios ("use-cases") using the domain layer; for example `CreateNoteService`, `UploadPDFService`.  
   - **Dependencies:**  
     - Uses the domain layer, utilizes interfaces for infrastructure (DB, file system, etc.).  
   - **Responsibility:**  
     - Arranges the application's business flows; applies domain rules using repository or external system interfaces.

3. **Interface Adapters (Adapter) Layer**  
   - **What it includes?**  
     - Infrastructure implementations such as database and file system that concretize the interfaces defined by the Use Case layer (e.g., `UserRepository`, `FileStorage`).  
     - For example, `adapter/postgres/userrepo.go`, `adapter/localfs/pdfstore.go`.  
   - **Dependencies:**  
     - External libraries (e.g., PostgreSQL driver, Redis client), OS calls, VPS API, etc. can be used in this layer.  
   - **Responsibility:**  
     - Managing database, file system, or external service communication; providing an abstract interface to the domain/use-case layer.

4. **Framework & Delivery (Infrastructure) Layer**  
   - **What it includes?**  
     - HTTP endpoints, routing, middleware, configuration, and logging with Go web framework (e.g., Fiber, Gin).  
   - **Dependencies:**  
     - The application runs by calling the Use Case layer. Interaction with the outside world (HTTP port, CLI, etc.) is defined here.  
   - **Responsibility:**  
     - Starting the application, handling user requests, and returning responses.

**Example Directory Structure:**

```
uninotes-backend/
├── cmd/
│   └── server/
│       └── main.go
├── domain/
│   ├── user.go
│   ├── note.go
│   └── ...
├── usecase/
│   ├── createnote.go
│   ├── uploadpdf.go
│   └── ...
├── adapter/
│   ├── postgres/
│   │   └── userrepo.go
│   ├── localfs/
│   │   └── pdfstore.go
│   └── cache/
│       └── redis.go
└── infrastructure/
    ├── http/
    │   └── router.go
    ├── env/
    │   └── config.go
    └── ...
```

---

## Database & Storage (VPS Focused)

1. **Repository Pattern**  
   - Interfaces like `UserRepository`, `NoteRepository`, `LikeRepository` in Go are defined with concrete implementations (CRUD, SQL queries) under `adapter/postgres/`.  
   - The `LikeRepository` implements a polymorphic approach to handle both note and PDF likes through a single model with content type differentiation.
   - Since the Use Case layer only calls interfaces, it becomes easier to switch from PostgreSQL to MySQL or a different storage system.

2. **File Storage**  
   - Instead of AWS S3, you can use the local file system or NFS share on a VPS.  
   - An implementation like `adapter/localfs/pdfstore.go` saves uploaded PDFs to a designated directory within the VPS.  
   - If high file traffic is expected, disk capacity and backup strategy need to be considered (e.g., mounting an additional disk, backup with rsync, etc.).

3. **Cache (Optional)**  
   - Redis or Memcached can be activated for performance boost in high-traffic requests.  
   - A file like `adapter/cache/redis.go` stores user sessions or frequently queried metadata.

---

## Service Layer (Use Case)

- **CreateNoteService**: Validates note creation requests from users, applies domain rules (e.g., title requirement, character limit), saves data through the repository, sends notifications if necessary.  
- **UploadPDFService**: Receives PDF file upload requests, forwards the file to the `pdfstore` adapter, keeps records in the database, can trigger tagging or note creation processes.  
- **LikeService**: Manages user interactions with content through likes, providing a unified approach for both notes and PDFs using a polymorphic content type system.
- **CollaborationService**: Manages versioning rules and real-time synchronization when multiple people need to edit a note.

These services work using domain entities (e.g., `Note`, `User`) and repository interfaces. Information flow is `Handler/Controller -> Use Case -> Repository -> DB`.

---

## Real-Time Collaboration

- Real-time editing between users is possible using **WebSockets** or SSE for applications hosted on a VPS.  
- You can manage each active connection in a separate thread with Go's concurrency model (goroutine, channel).  
- Advanced solutions like versioning (e.g., operational transform or CRDT) prevent conflicts in note content.

---

## Notification System

- **User Interaction:** Likes, comments, shared notes. The like system is optimized to handle both note and PDF content through a single model with content type differentiation.
- WebSocket or push notification (e.g., Firebase) can be used for real-time notifications.  
- If planning is desired (e.g., reminders like "upcoming exams"), a cron job or background queue system can be set up on the VPS.

---

## Security & Authorization

- **JWT** based authentication.  
- All sensitive data (password hash, token, etc.) is stored in the database; a field in `domain/user.go` might only store the hash of the password.  
- Basic security measures are taken at the VPS level with firewall tools like iptables/ufw. For example, only opening ports 80/443 and SSH (22) to the outside.  
- Free services like Let's Encrypt can be used for SSL certificates.

---

## Logging & Monitoring

- Application logs are transferred to a file within the VPS or to an external log collection service (e.g., Loki, ELK stack) in text or JSON format.  
- Monitoring tools like Prometheus/Grafana (request count, CPU/RAM usage, goroutine count) can be installed on the VPS or managed on an external server.  
- Load balancer configuration and horizontal scaling (adding additional VPS) should be considered when high traffic is reached.

---

## Deployment

1. **Environment Preparation:**  
   - Make sure Go runtime or Docker is installed on the VPS.  
   - Necessary port settings, firewall, SSH keys, SSL certificate.  

2. **CI/CD Pipeline:**  
   - Tests are run with GitHub Actions or GitLab CI, and if successful, deployment is made to the VPS via SSH or Docker Registry.  

3. **Backup & Scaling:**  
   - Data backups (PostgreSQL dump, syncing PDF files to another server with rsync).  
   - Plan to move to a higher capacity VPS plan or use multiple VPS (with load balancer) when traffic increases.

---

## Additional Development and Possible Extensions

- **Transition to Microservice Architecture:** When the system grows in the future, each module (note, user, interaction, etc.) can be separated into a separate service.  
- **Cache / Queue Based Approaches:** Background workers can be added for messaging (RabbitMQ, NATS) or ETL processes under heavy load.  
- **MinIO or S3 Compatible Storage:** It's possible to manage PDF files in a distributed and scalable way by setting up an S3-like object storage environment on a VPS or on a separate server.  

---

### Summary

This **systemPatterns.md** details how UniNotes can be structured in a **Go + Clean Architecture** and **VPS-based** infrastructure. Domain and Use Case layers are abstracted from infrastructural issues such as database and file system. The Repository pattern supports the strategy of storing PDFs in the local file system. Side services such as authentication, logging, and notification also become scalable and easily manageable by following this layered structure. VPS offers a simple and cost-effective solution for low/medium traffic in the initial phase, while it will be possible to transition to additional VPS or different cloud services as the project grows.
