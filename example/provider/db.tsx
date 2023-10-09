import { Level } from "level"
export const db = new Level<string, User>("./db", { valueEncoding: "json" })

interface User {
  Name: string
  ID: string
}
