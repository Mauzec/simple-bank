if [[ ! -f ./config/app.env ]]; then
  echo "❌ Error: Create a app.env file based on app.env.example"
  exit 1
fi

set -o allexport
source ./config/app.env
set +o allexport

export DB_SOURCE
export POSTGRES_USER
export POSTGRES_PASSWORD
export DB_PORT
export DB_NAME
export SSL_ENABLE
