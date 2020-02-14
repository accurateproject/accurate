
db = db.getSiblingDB('admin')
db.createUser(
  {
    user: "accurate",
    pwd: "accuRate",
    roles: [ { role: "userAdminAnyDatabase", db: "admin" } ]
  }
)
