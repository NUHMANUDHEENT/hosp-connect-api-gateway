# API Gateway Service

The **API Gateway Service** acts as the central entry point for all client requests in the HospConnect system. It simplifies client communication by managing API endpoints, routing requests to microservices via gRPC, and rendering a centralized HTML page for user-facing features.

---

## **Description**

This service provides:
- A unified endpoint for accessing various HospConnect microservices.
- Efficient communication using **gRPC** to handle requests and responses between services.
- A simple HTML page as the front-end interface for accessing certain features.

The gateway ensures secure, scalable, and efficient routing of traffic across all microservices.


---

## **Technology Stack**

- **Programming Language:** Go (Golang)  
- **Communication Protocols:** HTTP, gRPC  
- **Frameworks/Libraries:**  
  - **Gin**: For REST API endpoints and HTML rendering.  
  - **gRPC-Go**: For inter-service communication.  
  - **html/template**: For serving HTML pages.  
- **Security:**  
  - API key validation for REST endpoints.  
  - Mutual TLS for gRPC communication (optional).  

---

## **API Documentation**

### **REST Endpoints**

1. **Root API Entrypoint**
   - **URL:** `/api/v1`  
   - **Method:** `GET`  
   - **Description:** Returns metadata about the API Gateway and available services.

2. **User Management API Proxy**
   - **URL:** `/api/v1/users`  
   - **Description:** Fetch user details by forwarding the request to the **User Management Service**.  
   - **gRPC Endpoint Called:** `GetUser(userId)`  

3. **Doctor Managment API Proxy**
   - **URL:** `/api/v1/doctor`  
   - **Description:**Managing doctors details like sheduling, prescription managment etc..**.   

4. **Admin managment API Proxy**
   - **URL:** `/api/v1/admin`   
   - **Description:** Admin managment and patient , doctors managment.  

---

## **How to Set Up**

### **Clone the Repository**

```bash
git clone [https://github.com/your-username/api-gateway.git](https://github.com/NUHMANUDHEENT/hosp-connect-api-gateway.git)
cd api-gateway
