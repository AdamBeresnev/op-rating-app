# OP Rating app

TODO: The other half of the app
Uses Go, HTMX, and Tailwind CSS.

### Prerequisites

Install these first (yes, we still need node. Don't ask.):

1.  **Go**
2.  **Node.js and npm**
3.  **`golang-migrate/migrate` CLI**
    ```bash
    go install -tags 'sqlite3' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
    ```
4.  **`air`**
    ```bash
    go install github.com/air-verse/air@latest
    ```
5.  **TailwindCSS CLI**
    ```
    npm install tailwindcss @tailwindcss/cli
    ```

### Installation

1.  **Clone the repository:**

    ```bash
    git clone <your-repo-url>
    cd op-rating-app
    ```

2.  **Install frontend dependencies:**

    ```bash
    npm install
    ```

3.  **Set up the database:**
    ```bash
    migrate -database 'sqlite3://op_rating.db' -path migrations up
    ```

### Running the Development Server

You run both of these once, and when you make Go, HTML or CSS changes they automatically load the changes. You might need to refresh the browser though. Still very cool.

**Terminal 1: Start the Go Backend**

Use `air` to run the Go server with live-reloading.

```bash
air
```

**Terminal 2: Start the Tailwind CSS Watcher**

Re-compile your CSS automatically.

```bash
npx tailwindcss -i ./static/css/input.css -o ./static/css/output.css --watch
```

Once both are running, go to **http://localhost:8080**.

## Useful Commands

### Database Migrations

- **Apply all migrations:**
  ```bash
  migrate -database 'sqlite3://op_rating.db' -path migrations up
  ```
- **Revert the last migration:**
  ```bash
  migrate -database 'sqlite3://op_rating.db' -path migrations down
  ```

### Code Formatting

- **Format frontend files:**
  ```bash
  npm run format
  ```
- **Format Go files:**
  ```bash
  gofmt -w .
  ```
