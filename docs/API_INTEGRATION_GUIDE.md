# Root Revolution — Frontend API Integration Guide

**Version:** 1.0.0
**Base URL:** `https://cloudcalls.easipath.com/backend_rootrevolution/api`
**App Name:** `rootrevolutionapi`
**Content-Type:** `application/json`

---

## Table of Contents

1. [Overview](#overview)
2. [Authentication](#authentication)
3. [Product Endpoints](#product-endpoints)
4. [User & Auth Endpoints](#user--auth-endpoints)
5. [Data Models](#data-models)
6. [Error Handling](#error-handling)
7. [Product Change Approval Flow](#product-change-approval-flow)
8. [Image Upload (Base64)](#image-upload-base64)
9. [Code Examples](#code-examples)

---

## Overview

### Access Levels

| Access | Description |
|--------|-------------|
| **Public** | No authentication required |
| **Private** | Requires a valid JWT Bearer token |
| **Admin** | Requires JWT + admin role |

### Base URL
```
https://cloudcalls.easipath.com/backend_rootrevolution/api
```

---

## Authentication

The API uses **JWT Bearer tokens**. Obtain a token by logging in, then include it in all private request headers.

### Header Format
```
Authorization: Bearer <your_token_here>
```

### Token Lifetime
Tokens expire after **24 hours**. Re-login to get a new token.

---

## Product Endpoints

### 1. List All Products
**`GET /products`** — Public

Returns all active products. Optionally filter by category.

**Query Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `category` | string | No | Filter by category name |

**Example Requests**
```
GET https://cloudcalls.easipath.com/backend_rootrevolution/api/products
GET https://cloudcalls.easipath.com/backend_rootrevolution/api/products?category=Herbal%20and%20natural%20supplements
```

**Success Response — `200 OK`**
```json
{
  "data": [
    {
      "app_name": "rootrevolutionapi",
      "org": "C10201",
      "id": 20101,
      "name": "120 Moringa Capsules",
      "description": "Experience the power of nature...",
      "category": "Herbal and natural supplements",
      "image": "https://i.postimg.cc/W3TTLqLJ/Moringa-120-Capsules.png",
      "price": 150,
      "original_price": 150,
      "discount_percentage": 0,
      "stock_quantity": 50,
      "is_new": true,
      "is_best_seller": true,
      "is_on_sale": true,
      "created_at": "2025-05-28T16:08:27Z",
      "updated_at": "0001-01-01T00:00:00Z",
      "created_by": "test@home.com",
      "updated_by": "",
      "status": "active"
    }
  ],
  "total": 33
}
```

---

### 2. Get Single Product
**`GET /products/:id`** — Public

**URL Parameters**

| Parameter | Type | Description |
|-----------|------|-------------|
| `id` | integer | Product ID (e.g. `20101`) |

**Example Request**
```
GET https://cloudcalls.easipath.com/backend_rootrevolution/api/products/20101
```

**Success Response — `200 OK`**
```json
{
  "data": {
    "app_name": "rootrevolutionapi",
    "org": "C10201",
    "id": 20101,
    "name": "120 Moringa Capsules",
    "description": "Experience the power of nature...",
    "category": "Herbal and natural supplements",
    "image": "https://i.postimg.cc/W3TTLqLJ/Moringa-120-Capsules.png",
    "price": 150,
    "original_price": 150,
    "discount_percentage": 0,
    "stock_quantity": 50,
    "is_new": true,
    "is_best_seller": true,
    "is_on_sale": true,
    "created_at": "2025-05-28T16:08:27Z",
    "updated_at": "0001-01-01T00:00:00Z",
    "created_by": "test@home.com",
    "updated_by": "",
    "status": "active"
  }
}
```

**Error Response — `404 Not Found`**
```json
{
  "error": "Product not found"
}
```

---

### 3. Create Product
**`POST /products`** — Private

> **Important:** Creating a product does **not** apply immediately. The system saves the request and sends an authorization email to the owner. The product is only created once the owner approves the link.

**Request Headers**
```
Authorization: Bearer <token>
Content-Type: application/json
```

**Request Body**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | **Yes** | Product name |
| `description` | string | No | Full product description |
| `category` | string | No | Product category |
| `image` | string | No | Image URL **or** base64-encoded image (auto-uploaded to Dropbox) |
| `price` | number | **Yes** | Current price (must be > 0) |
| `original_price` | number | No | Original price before discount |
| `discount_percentage` | number | No | Discount % (0–100) |
| `stock_quantity` | integer | No | Available stock count |
| `is_new` | boolean | No | Mark as new product |
| `is_best_seller` | boolean | No | Mark as best seller |
| `is_on_sale` | boolean | No | Mark as on sale |
| `status` | string | No | `"active"` or `"inactive"` (defaults to `"active"`) |

**Example Request Body**
```json
{
  "name": "50g Turmeric Powder",
  "description": "Pure turmeric powder with high curcumin content.",
  "category": "Herbal and natural supplements",
  "image": "https://example.com/turmeric.png",
  "price": 75,
  "original_price": 75,
  "discount_percentage": 0,
  "stock_quantity": 100,
  "is_new": true,
  "is_best_seller": false,
  "is_on_sale": false,
  "status": "active"
}
```

**Success Response — `202 Accepted`**
```json
{
  "message": "Product creation request submitted. An authorization email has been sent to the owner for approval.",
  "token": "550e8400-e29b-41d4-a716-446655440000"
}
```

> Save the `token` — you can use it to track the request status.

---

### 4. Update Product
**`PUT /products/:id`** — Private

> **Important:** All fields are **optional** — only send the fields you want to change. The update does not apply immediately; an authorization email is sent to the owner first.

**URL Parameters**

| Parameter | Type | Description |
|-----------|------|-------------|
| `id` | integer | Product ID to update |

**Request Headers**
```
Authorization: Bearer <token>
Content-Type: application/json
```

**Request Body** *(all fields optional — send only what changes)*

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | New product name |
| `description` | string | New description |
| `category` | string | New category |
| `image` | string | New image URL or base64 image |
| `price` | number | New price |
| `original_price` | number | New original price |
| `discount_percentage` | number | New discount % |
| `stock_quantity` | integer | New stock quantity |
| `is_new` | boolean | Update new flag |
| `is_best_seller` | boolean | Update best seller flag |
| `is_on_sale` | boolean | Update on-sale flag |
| `status` | string | `"active"` or `"inactive"` |

**Example — Update price and stock only**
```json
{
  "price": 140,
  "stock_quantity": 75,
  "is_on_sale": true
}
```

**Example — Update with base64 image**
```json
{
  "name": "120 Moringa Capsules (New Label)",
  "image": "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAA..."
}
```

**Success Response — `202 Accepted`**
```json
{
  "message": "Product update request submitted. An authorization email has been sent to the owner for approval.",
  "token": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Error Response — `404 Not Found`**
```json
{
  "error": "product 20101 not found"
}
```

---

### 5. Delete Product
**`DELETE /products/:id`** — Private

> **Important:** Deletion also requires owner authorization via email before it takes effect.

**URL Parameters**

| Parameter | Type | Description |
|-----------|------|-------------|
| `id` | integer | Product ID to delete |

**Request Headers**
```
Authorization: Bearer <token>
```

**Example Request**
```
DELETE https://cloudcalls.easipath.com/backend_rootrevolution/api/products/20101
```

**Success Response — `202 Accepted`**
```json
{
  "message": "Product deletion request submitted. An authorization email has been sent to the owner for approval.",
  "token": "550e8400-e29b-41d4-a716-446655440000"
}
```

---

### 6. Authorize Product Change *(Owner Only)*
**`GET /products/authorize/:token`** — Public (email link)

This endpoint is called automatically when the owner clicks the authorization link in the approval email. It is **not** intended to be called directly by the frontend — it returns an HTML confirmation page.

**URL Parameters**

| Parameter | Type | Description |
|-----------|------|-------------|
| `token` | UUID string | Authorization token from the email |

**Success Response** — HTML page confirming the change was applied.

**Error Response** — HTML page with the reason for failure (expired, already used, invalid).

> **Token expiry:** Authorization links expire after **24 hours**.

---

## User & Auth Endpoints

### 7. Login
**`POST /auth/login`** — Public

**Request Body**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `email` | string | **Yes** | User email |
| `password` | string | **Yes** | User password |

**Example Request Body**
```json
{
  "email": "biangacila@gmail.com",
  "password": "Nathan010309*"
}
```

**Success Response — `200 OK`**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "name": "Merveilleux",
    "surname": "Biangacila",
    "email": "biangacila@gmail.com",
    "role": "admin",
    "status": "active"
  }
}
```

**Error Response — `401 Unauthorized`**
```json
{
  "error": "invalid credentials"
}
```

---

### 8. Get Current User (Me)
**`GET /auth/me`** — Private

Returns the currently authenticated user's profile decoded from the JWT.

**Request Headers**
```
Authorization: Bearer <token>
```

**Success Response — `200 OK`**
```json
{
  "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "email": "biangacila@gmail.com",
  "role": "admin",
  "name": "Merveilleux",
  "surname": "Biangacila"
}
```

---

### 9. Register New User
**`POST /auth/register`** — Admin Only

**Request Headers**
```
Authorization: Bearer <admin_token>
Content-Type: application/json
```

**Request Body**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | **Yes** | First name |
| `surname` | string | No | Last name |
| `email` | string | **Yes** | Email address (must be unique) |
| `password` | string | **Yes** | Password |
| `role` | string | No | `"admin"`, `"editor"`, or `"viewer"` (default: `"editor"`) |

**Example Request Body**
```json
{
  "name": "Jane",
  "surname": "Doe",
  "email": "jane@example.com",
  "password": "SecurePass123!",
  "role": "editor"
}
```

**Success Response — `201 Created`**
```json
{
  "message": "User registered successfully",
  "user": {
    "id": "b2c3d4e5-f6a7-8901-bcde-f12345678901",
    "name": "Jane",
    "surname": "Doe",
    "email": "jane@example.com",
    "role": "editor"
  }
}
```

---

### 10. List Users
**`GET /users`** — Admin Only

**Request Headers**
```
Authorization: Bearer <admin_token>
```

**Success Response — `200 OK`**
```json
{
  "data": [
    {
      "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
      "name": "Merveilleux",
      "surname": "Biangacila",
      "email": "biangacila@gmail.com",
      "role": "admin",
      "status": "active",
      "created_at": "2026-04-01T11:00:00Z"
    }
  ],
  "total": 1
}
```

---

### 11. Get User by ID
**`GET /users/:id`** — Admin Only

**URL Parameters**

| Parameter | Type | Description |
|-----------|------|-------------|
| `id` | UUID string | User ID |

---

### 12. Update User
**`PUT /users/:id`** — Admin Only

**Request Body** *(all optional)*

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | New first name |
| `surname` | string | New last name |
| `role` | string | New role: `"admin"`, `"editor"`, `"viewer"` |
| `status` | string | `"active"` or `"inactive"` |

**Success Response — `200 OK`**
```json
{
  "message": "User updated successfully",
  "data": { ... }
}
```

---

### 13. Delete User
**`DELETE /users/:id`** — Admin Only

**Success Response — `200 OK`**
```json
{
  "message": "User deleted successfully"
}
```

---

### 14. Health Check
**`GET /health`** — Public

```
GET https://cloudcalls.easipath.com/backend_rootrevolution/api/health
```

**Response — `200 OK`**
```json
{
  "status": "ok",
  "app": "rootrevolutionapi",
  "version": "1.0.0"
}
```

---

## Data Models

### Product Object

```json
{
  "app_name":            "rootrevolutionapi",
  "org":                 "C10201",
  "id":                  20101,
  "name":                "120 Moringa Capsules",
  "description":         "Full product description...",
  "category":            "Herbal and natural supplements",
  "image":               "https://...",
  "price":               150.00,
  "original_price":      150.00,
  "discount_percentage": 0.00,
  "stock_quantity":      50,
  "is_new":              true,
  "is_best_seller":      true,
  "is_on_sale":          true,
  "created_at":          "2025-05-28T16:08:27Z",
  "updated_at":          "0001-01-01T00:00:00Z",
  "created_by":          "user@example.com",
  "updated_by":          "",
  "status":              "active"
}
```

### User Object

```json
{
  "id":         "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "name":       "Merveilleux",
  "surname":    "Biangacila",
  "email":      "biangacila@gmail.com",
  "role":       "admin",
  "status":     "active",
  "created_at": "2026-04-01T11:00:00Z"
}
```

### User Roles

| Role | Description |
|------|-------------|
| `admin` | Full access — manage products, users, all operations |
| `editor` | Can request product changes (still requires owner approval) |
| `viewer` | Read-only via authenticated routes |

### Product Status Values

| Value | Description |
|-------|-------------|
| `active` | Visible and available |
| `inactive` | Hidden from public listing |

---

## Error Handling

All error responses follow this structure:

```json
{
  "error": "Human-readable error message"
}
```

### HTTP Status Codes

| Code | Meaning |
|------|---------|
| `200` | Success |
| `201` | Created |
| `202` | Accepted (change submitted, awaiting owner approval) |
| `400` | Bad request — invalid input or parameters |
| `401` | Unauthorized — missing or invalid JWT token |
| `403` | Forbidden — insufficient role (e.g. non-admin) |
| `404` | Not found — resource does not exist |
| `500` | Internal server error |

---

## Product Change Approval Flow

All write operations (create, update, delete) go through an **owner approval flow**:

```
Frontend                     API Server                    Owner Email
   |                             |                              |
   |--- POST /products --------> |                              |
   |                             |-- Save pending request       |
   |                             |-- Send approval email -----> |
   |<-- 202 Accepted + token --- |                              |
   |                             |                              |
   |                             |          (owner clicks link) |
   |                             | <--- GET /authorize/:token --|
   |                             |-- Apply change to database   |
   |                             |-- Return success HTML -----> |
```

**What this means for the frontend:**
- A `202 Accepted` response means the request was **received and queued**, not applied.
- Show the user a message like: *"Your change has been submitted for approval. You will be notified once it is applied."*
- The owner receives an email and clicks **Authorize Change** — only then is the product updated.
- Links expire in **24 hours**.

---

## Image Upload (Base64)

When submitting a product image as a **base64-encoded string**, the API automatically uploads it to Dropbox and stores the resulting CDN URL.

### Supported formats
`image/jpeg`, `image/jpg`, `image/png`, `image/gif`, `image/webp`

### How to send

**Option A — Data URI format (recommended)**
```json
{
  "image": "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwADhQGAWjR9awAAAABJRU5ErkJggg=="
}
```

**Option B — URL (no upload, used as-is)**
```json
{
  "image": "https://example.com/images/product.png"
}
```

The API auto-detects which format you've sent and handles it appropriately.

---

## Code Examples

### JavaScript / Fetch

#### Login and store token
```javascript
const login = async () => {
  const res = await fetch(
    'https://cloudcalls.easipath.com/backend_rootrevolution/api/auth/login',
    {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        email: 'biangacila@gmail.com',
        password: 'Nathan010309*'
      })
    }
  );
  const data = await res.json();
  localStorage.setItem('token', data.token);
  return data;
};
```

#### Get all products (public)
```javascript
const getProducts = async (category = '') => {
  const url = new URL('https://cloudcalls.easipath.com/backend_rootrevolution/api/products');
  if (category) url.searchParams.set('category', category);

  const res = await fetch(url);
  const data = await res.json();
  return data.data; // array of products
};
```

#### Get single product
```javascript
const getProduct = async (id) => {
  const res = await fetch(
    `https://cloudcalls.easipath.com/backend_rootrevolution/api/products/${id}`
  );
  const data = await res.json();
  return data.data;
};
```

#### Create product (with auth)
```javascript
const createProduct = async (product) => {
  const token = localStorage.getItem('token');
  const res = await fetch(
    'https://cloudcalls.easipath.com/backend_rootrevolution/api/products',
    {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify(product)
    }
  );
  return res.json();
  // { message: "...", token: "uuid" }
};
```

#### Update product (partial update)
```javascript
const updateProduct = async (id, changes) => {
  const token = localStorage.getItem('token');
  const res = await fetch(
    `https://cloudcalls.easipath.com/backend_rootrevolution/api/products/${id}`,
    {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify(changes)
    }
  );
  return res.json();
};

// Example: update only price
await updateProduct(20101, { price: 140, is_on_sale: true });
```

#### Update product with base64 image
```javascript
const updateProductImage = async (id, imageFile) => {
  const token = localStorage.getItem('token');

  // Convert file to base64
  const base64 = await new Promise((resolve) => {
    const reader = new FileReader();
    reader.onload = (e) => resolve(e.target.result);
    reader.readAsDataURL(imageFile); // produces "data:image/png;base64,..."
  });

  const res = await fetch(
    `https://cloudcalls.easipath.com/backend_rootrevolution/api/products/${id}`,
    {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify({ image: base64 })
    }
  );
  return res.json();
};
```

#### Delete product
```javascript
const deleteProduct = async (id) => {
  const token = localStorage.getItem('token');
  const res = await fetch(
    `https://cloudcalls.easipath.com/backend_rootrevolution/api/products/${id}`,
    {
      method: 'DELETE',
      headers: { 'Authorization': `Bearer ${token}` }
    }
  );
  return res.json();
};
```

---

### React Hook (reusable)

```javascript
import { useState, useCallback } from 'react';

const BASE_URL = 'https://cloudcalls.easipath.com/backend_rootrevolution/api';

export const useRootRevolutionAPI = () => {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  const getToken = () => localStorage.getItem('rr_token');

  const authHeaders = () => ({
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${getToken()}`
  });

  const request = useCallback(async (path, options = {}) => {
    setLoading(true);
    setError(null);
    try {
      const res = await fetch(`${BASE_URL}${path}`, options);
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || 'Request failed');
      return data;
    } catch (err) {
      setError(err.message);
      throw err;
    } finally {
      setLoading(false);
    }
  }, []);

  const login = useCallback(async (email, password) => {
    const data = await request('/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password })
    });
    localStorage.setItem('rr_token', data.token);
    return data;
  }, [request]);

  const getProducts = useCallback((category = '') => {
    const qs = category ? `?category=${encodeURIComponent(category)}` : '';
    return request(`/products${qs}`);
  }, [request]);

  const getProduct = useCallback((id) => request(`/products/${id}`), [request]);

  const createProduct = useCallback((product) => request('/products', {
    method: 'POST',
    headers: authHeaders(),
    body: JSON.stringify(product)
  }), [request]);

  const updateProduct = useCallback((id, changes) => request(`/products/${id}`, {
    method: 'PUT',
    headers: authHeaders(),
    body: JSON.stringify(changes)
  }), [request]);

  const deleteProduct = useCallback((id) => request(`/products/${id}`, {
    method: 'DELETE',
    headers: authHeaders()
  }), [request]);

  return {
    loading,
    error,
    login,
    getProducts,
    getProduct,
    createProduct,
    updateProduct,
    deleteProduct
  };
};
```

---

### Vue.js Composable

```javascript
import { ref } from 'vue';

const BASE_URL = 'https://cloudcalls.easipath.com/backend_rootrevolution/api';

export function useRootRevolutionAPI() {
  const loading = ref(false);
  const error = ref(null);

  const getToken = () => localStorage.getItem('rr_token');
  const authHeaders = () => ({
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${getToken()}`
  });

  const request = async (path, options = {}) => {
    loading.value = true;
    error.value = null;
    try {
      const res = await fetch(`${BASE_URL}${path}`, options);
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || 'Request failed');
      return data;
    } catch (err) {
      error.value = err.message;
      throw err;
    } finally {
      loading.value = false;
    }
  };

  return {
    loading,
    error,
    login: (email, password) => request('/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password })
    }),
    getProducts: (category = '') => {
      const qs = category ? `?category=${encodeURIComponent(category)}` : '';
      return request(`/products${qs}`);
    },
    getProduct: (id) => request(`/products/${id}`),
    createProduct: (product) => request('/products', {
      method: 'POST',
      headers: authHeaders(),
      body: JSON.stringify(product)
    }),
    updateProduct: (id, changes) => request(`/products/${id}`, {
      method: 'PUT',
      headers: authHeaders(),
      body: JSON.stringify(changes)
    }),
    deleteProduct: (id) => request(`/products/${id}`, {
      method: 'DELETE',
      headers: authHeaders()
    })
  };
}
```

---

## Quick Reference

| Method | Endpoint | Access | Description |
|--------|----------|--------|-------------|
| `GET` | `/health` | Public | Health check |
| `GET` | `/products` | Public | List all products |
| `GET` | `/products/:id` | Public | Get product by ID |
| `GET` | `/products/authorize/:token` | Public | Apply approved change (email link) |
| `POST` | `/products` | Private | Submit new product (requires approval) |
| `PUT` | `/products/:id` | Private | Submit product update (requires approval) |
| `DELETE` | `/products/:id` | Private | Submit product deletion (requires approval) |
| `POST` | `/auth/login` | Public | Login, get JWT token |
| `GET` | `/auth/me` | Private | Get current user info |
| `POST` | `/auth/register` | Admin | Register new user |
| `GET` | `/users` | Admin | List all users |
| `GET` | `/users/:id` | Admin | Get user by ID |
| `PUT` | `/users/:id` | Admin | Update user |
| `DELETE` | `/users/:id` | Admin | Delete user |

---

*For technical issues or integration support, contact the backend team.*
*API hosted at: `https://cloudcalls.easipath.com`*
