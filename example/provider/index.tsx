import express, { Request } from "express"
import { Fragment, h } from "preact" // required imports
import React from "preact/compat" // required import (or use other option in tsconfig.json)
import { render } from "preact-render-to-string" // render to string
import HomePage, { UserIDCookieName } from "./pages/Home"
import cookieParser from "cookie-parser"
import "./db"
import { db } from "./db"
import ConsentScreen from "./pages/ConsentScreen"

declare global {
  // namespace Express {
  //   interface Request {
  //     id: string
  //   }
  // }

  namespace NodeJS {
    interface ProcessEnv {
      CW_API: string
      CW_ADMIN_KEY: string
    }
  }
}

const app = express()
const port = 8080
app.use(cookieParser())
app.use(express.urlencoded())
app.use((req, res, next) => {
  console.log("request: ", req.method, req.path, req.query, req.body)
  next()
})

app.get("/", async (req, res) => {
  res.set("Content-Type", "text/html")
  res.send(await HomePage(req, res))
})

app.post("/register", async (req: Request<{}, {}, { name: string }>, res) => {
  res.set("Content-Type", "text/html")
  const userID = crypto.randomUUID()
  await db.put(userID, {
    Name: req.body.name,
    ID: userID,
  })
  res.cookie(UserIDCookieName, userID)
  res.set("HX-Redirect", "/") // special header to force redirect, see https://htmx.org/reference/#response_headers
  res.send()
})

app.get(
  "/consent",
  async (req: Request<{}, {}, {}, { client_id: string }>, res) => {
    await ConsentScreen(req, res)
  }
)

app.listen(port, () => {
  console.log(`Listening on port ${port}...`)
})
