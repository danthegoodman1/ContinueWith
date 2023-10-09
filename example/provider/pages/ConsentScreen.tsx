import { Request, Response } from "express"
import React from "preact/compat"
import { render } from "preact-render-to-string"
import Body from "./Body"
import { UserIDCookieName } from "./Home"
import { db } from "../db"
import { Client, GetClient } from "../continue_with"

export default async function ConsentScreen(
  req: Request<{}, {}, {}, { client_id: string }>,
  res: Response
) {
  res.set("Content-Type", "text/html")
  const cookies = req.cookies
  const userID: string | undefined = cookies[UserIDCookieName]
  console.log("user id", userID)

  if (!userID) {
    return res.redirect("/")
  }

  const user = await db.get(userID)
  if (!user) {
    res.clearCookie(UserIDCookieName)
    return res.redirect("/")
  }

  // Get the client info from ContinueWith
  let client: Client
  try {
    client = await GetClient(req.query.client_id)
  } catch (error) {
    console.error(error)
    return res.send(
      render(
        <Body>
          <h1>I BROKE!</h1>
          <p>Was that a valid client id?</p>
        </Body>
      )
    )
  }

  return res.send(
    render(
      <Body>
        <h1>Log in with X</h1>
      </Body>
    )
  )
}
