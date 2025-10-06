

### Create new user
POST /users

Body:
```json
{
    "username": "some",
    "password": "somepassword",
    "full_name": "Some Mone",
    "email": "some@some.some"
}
```

Returns user object if ok



### Login user
POST /users/login

Body:
```json
{
    "username": "some",
    "password": "somepassword"
}
```

Returns:
```json
{
    "token": "<symmetric PASETO token>",
    "user": "<user object>"
}
```



### Create new account (for current user)
POST /accounts

Headers:
- `Authorization: Bearer <token>`

Body:
```json
{
    "currency": "USD" // or EUR, KZT, RUB, BYN, GBP
}
```

Returns created account object if ok



### Get account by id (for current user)
GET /accounts/:id

Headers:
- `Authorization: Bearer <token>`

Path param:
- `id` (int64)

Returns account object if ok



### Get list of accounts (for current user)
GET /accounts

Headers:
- `Authorization: Bearer <token>`

Query params:
- `page_id` (int32, min=1)
- `page_size` (int32, min=5, max=30)

Returns array of accounts (for current user) if ok



### Update balance (available for current user)
PUT /accounts

Headers:
- `Authorization: Bearer <token>`

Body:
```json
{
    "id": 1,
    "balance": 100.0
}
```

Returns updated account object if ok



### Delete account (available for current user)
DELETE /accounts/:id

Notes:
- Cannot delete if there are transfers with this account

Headers:
- `Authorization: Bearer <token>`

Path param:
- `id` (int64)

Returns `200 OK` if ok



### Create transfer between accounts (from current user to any)
POST /transfers

Headers:
- `Authorization: Bearer <token>`

Body:
```json
{
    "from_account_id": 1,
    "to_account_id": 2,
    "amount": 10.0,
    "currency": "USD" // or EUR, KZT, RUB, BYN, GBP
}
```

Returns transfer object if ok
