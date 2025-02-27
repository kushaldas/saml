import json

password = "$2a$10$mk8JBytwtLRB4qUB3jQnEeIyTh28KxNI03ijqB6LBSz7ZgExYsJSq"
users = []
user1 = {
    "Password" : password,
    "Groups": ["Administrators", "Users"],
    "Email": "alice@example.se",
    "common_name": "Alice Smith",
    "Surname": "Smith",
    "given_name": "Alice",
    "name": "alice"
}
user2 = {
    "Password" : password,
    "Groups": [ "Users",],
    "Email": "bob@example.se",
    "common_name": "Bob Smith",
    "Surname": "Smith",
    "given_name": "Bob",
    "name": "bob"
}

users.append(user1)
users.append(user2)
with open("./users.json", "w") as f:
    json.dump(users, f)
