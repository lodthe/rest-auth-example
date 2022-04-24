# rest-auth-example

An example of REST API with authentication (via JWT). 
The server provides API for registration and some other CRUDs.

# API

This section describes how to use the API. Also, there is a [postman collection](./postman-collection.json).

## Registration

Register a new user.

```
POST /api/auth/register
-> {
    "username": "random_user",
    "email": "wowowow.gmail@gmail.com", 
    "avatar": "https://images.google.com/"
    "sex": "male"
}


<- {
    "result": {
        "refresh_token": "eyJhbG...refresh"
    }
}
```
## Issue an access-token

Provide `Authorization: Bearer <refresh token>` header to issue a new access-token:

```
POST /api/auth/issue-access-token
Authorization: Bearer eyJhbG...refresh

<- {
    "result": {
        "access_token": "eyJhbG...access"
    }
}
```

All the following requests should be sent with an issued access-token.

## Get your profile

```
GET /api/users/myself
Authorization: Bearer eyJhbG...access

<- {
    "result": {
        "id": "4cb81bf5-4520-4861-85d2-ec7ceb744115",
        "username": "xXx__WINNER__xXx",
        "sex": "male",
        "email": "wowowow.gmail@gmail.com"
    }
}
```

## Update your profile

```
PUT /api/users/myself
Authorization: Bearer eyJhbG...access
-> {
    "username": "xXx_MAFIOZI_xXx",
    "email": "soa.enjoyer@gmail.com",
    "avatar": "https://www.hollywoodreporter.com/wp-content",
    "sex": "exmale"
}

<- {
    "result": {
        "id": "4cb81bf5-4520-4861-85d2-ec7ceb744115",
        "username": "xXx_MAFIOZI_xXx",
        "avatar": "https://www.hollywoodreporter.com/wp-content",
        "sex": "exmale",
        "email": "soa.enjoyer@gmail.com"
    }
}
```

## Get profiles by usernames

```
GET /api/users?usernames=random_user,xXx__WINNER__xXx
Authorization: Bearer eyJhbG...access

<- {
    "result": {
        "users": [
            {
                "id": "4cb81bf5-4520-4861-85d2-ec7ceb744115",
                "username": "xXx__WINNER__xXx",
                "sex": "male",
                "email": "wowowow.gmail@gmail.com"
            },
            {
                "id": "d0ed4202-ea84-4c38-b89a-35830fcaa335",
                "username": "random_user",
                "sex": "male",
                "email": "wowowow.gmail@gmail.com"
            }
        ]
    }
}
```

## Create stats task

Create a task to get a users' statistics asynchronously. 

The server will send the request to RabbitMQ. A worker consumes requests and uploads generated stats-documents to YandexCloud S3.

```
POST /api/stats/xXx__WINNER__xXx
Authorization: Bearer eyJhbG...access
        
<- {
    "result": {
        "id": "179c089a-827e-4436-a251-843131baa1e0"
    }
}
```

## Check stats-task status

```
GET /api/stats/tasks/179c089a-827e-4436-a251-843131baa1e0
Authorization: Bearer eyJhbG...access
        
<- {
    "result": {
        "id": "179c089a-827e-4436-a251-843131baa1e0",
        "status": "DONE",
        "document_url": "https://storage.yandexcloud.net/soa-stats/stats-179c089a-827e-4436-a251-843131baa1e0.pdf"
    }
}
```

If the status is DONE, there is a link to the generated document is given.

