# Go REST API Gateway for Next.js Frontend

Next.js communicates exclusively through the Go Backend REST API rather than querying Supabase directly via frontend clients. We chose this architecture to centralize domain queries, score filtering, and authorization logic within the Go service while keeping Next.js strictly focused on rendering and UI interaction.
