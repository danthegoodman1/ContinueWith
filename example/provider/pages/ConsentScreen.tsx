import { Request, Response } from "express"
import React from "preact/compat"
import { render } from "preact-render-to-string"
import Body from "./Body"
import { UserIDCookieName } from "./Home"
import { db } from "../db"
import { Client, GetClient } from "../continue_with"

export interface ConsentParams {
  /**
   * This will be `code`
   */
  response_type: string
  client_id: string
  redirect_uri: string
  /**
   * Need to split by spaces
   */
  scope: string
  state?: string
}

export default async function ConsentScreen(
  req: Request<{}, {}, {}, ConsentParams>,
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
        <h1>Log in with {client.Name}?</h1>
        <h3>Scopes:</h3>
        <p>{"(typically you'd just explain each one)"}</p>
        <ul>
          {req.query.scope
            .split(" ")
            .filter((scope) => scope !== "")
            .map((scope) => {
              return <li>{scope}</li>
            })}
        </ul>
        <button hx-post="/consent-approve">Approve</button>
        <button hx-post="/consent-deny">Deny</button>
      </Body>
    )
  )
}
