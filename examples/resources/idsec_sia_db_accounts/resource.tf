
# PAM Account
resource "idsec_sia_strong_accounts" "pam_account" {
  store_type   = "pam"
  name         = "MyPAMAccount"
  account_name = var.account_name
  safe         = var.safe_name
}

# PostgreSQL
# Managed PostgreSQL database account
resource "idsec_sia_strong_accounts" "postgresql" {
  store_type = "managed"
  name       = "PostgreSQL_Production_DB"
  address    = "postgres.example.com"
  database   = "production_db"
  platform   = "PostgreSQL"
  port       = 5432
  username   = var.username
  password   = var.password
}

# MySQL
# Managed MySQL database account
resource "idsec_sia_strong_accounts" "mysql" {
  store_type = "managed"
  name       = "MySQL_App_DB"
  address    = "mysql.example.com"
  database   = "app_database"
  platform   = "MySQL"
  port       = 3306
  username   = var.username
  password   = var.password
}

# MariaDB
# Managed MariaDB database account
resource "idsec_sia_strong_accounts" "mariadb" {
  store_type = "managed"
  name       = "MariaDB_Analytics"
  address    = "mariadb.example.com"
  database   = "analytics_db"
  platform   = "MariaDB"
  port       = 3306
  username   = var.username
  password   = var.password
}

# Microsoft SQL Server
# Managed Microsoft SQL Server account
resource "idsec_sia_strong_accounts" "mssql" {
  store_type = "managed"
  name       = "MSSQL_Enterprise"
  address    = "sqlserver.example.com"
  database   = "EnterpriseDB"
  platform   = "MSSql"
  port       = 1433
  username   = var.username
  password   = var.password
}

# Oracle Database
# Managed Oracle database account
resource "idsec_sia_strong_accounts" "oracle" {
  store_type = "managed"
  name       = "Oracle_ERP_System"
  address    = "oracle.example.com"
  database   = "ORCL"
  platform   = "Oracle"
  port       = 1521
  username   = var.username
  password   = var.password
}


# MongoDB
# Managed MongoDB database account (requires address and database)
resource "idsec_sia_strong_accounts" "mongodb" {
  store_type    = "managed"
  name          = "MongoDB_DocStore"
  address       = "mongodb.example.com"
  auth_database = "admin"
  database      = "documents"
  platform      = "MongoDB"
  port          = 27017
  username      = var.username
  password      = var.password
}

# DB2 Unix SSH
# Managed DB2 Unix SSH account (requires address)
resource "idsec_sia_strong_accounts" "db2_unix_ssh" {
  store_type = "managed"
  name       = "DB2_Mainframe"
  address    = "db2server.example.com"
  platform   = "DB2UnixSSH"
  username   = var.username
  password   = var.password
}

# Windows Domain
# Managed Windows Domain account (requires address)
resource "idsec_sia_strong_accounts" "windows_domain" {
  store_type = "managed"
  name       = "WinDomain_ServiceAccount"
  address    = "dc.example.com"
  platform   = "WinDomain"
  username   = var.username
  password   = var.password
}

# AWS Access Keys
# AWS IAM user access keys (uses secretAccessKey instead of password)
resource "idsec_sia_strong_accounts" "aws_access_keys" {
  store_type        = "managed"
  name              = "AWS_Production_User"
  aws_access_key_id = var.iam_access_key_id
  aws_account_id    = var.iam_account
  platform          = "AWSAccessKeys"
  username          = var.iam_username
  secret_access_key = var.iam_secret_access_key
}