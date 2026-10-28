import { Pool } from "pg";
import { drizzle } from "drizzle-orm/node-postgres";
import * as schema from "./schema";

// SSL follows the connection string: add sslmode=require for hosted databases like Neon,
// leave it off for a Postgres without SSL (e.g. on the Dokploy server).
const pool = new Pool({
  connectionString: process.env.DATABASE_URL,
});

export const db = drizzle(pool, { schema });
