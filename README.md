# Description
The service is already running on my droplet, but if you want to deploy/test it yourself go to "Usage"

## 1. How many Twitter users are in the database?
Endpoint: /users
```json
{
    "users": 659774
}
```

## 2. Which Twitter users link the most to other Twitter users? (Provide the top ten.)
Endpoint: /mentioners
```json
[
    {
    "_id": {
        "user": "lost_dog"
    },
    "count": 549
    },
    ...
]
```
## 3. Who is are the most mentioned Twitter users? (Provide the top five.)
Endpoint: /mentioned
```json
[
    {
        "User": "mileycyrus",
        "Count": 4500
    },
    ...
]
```
## 4. Who are the most active Twitter users (top ten)?
Endpoint: /active
```json
[
    {
        "_id": {
            "user":"lost_dog"
        },
        "count": 549
    },
    ...
]
```
## 5. Who are the five most grumpy (most negative tweets) and the most happy (most positive tweets)? (Provide five users for each group)
Endpoint: /polarity
```json
{
    "negative": [
        {
        "_id": {
            "user": "lost_dog"
        },
        "count": 549
        },
        ...
    ],
    "positive": [
        {
        "_id": {
            "user": "what_bugs_u"
        },
        "count": 246
        },
        ...
    ]
}
```

# Usage
The database can be started with:
```bash
docker run --rm -d borum/my-mongodb
```

The app can be deployed by executing:
```bash
docker run --rm -d -p 80:8080 borum/mongo-app
```

when both containers are up and running, you can reach the endpoints on your localhost (port 80)
