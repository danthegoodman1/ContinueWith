import { Request, Response } from 'express'
import render from 'preact-render-to-string'
import React from 'preact/compat'
import { Fragment, h } from 'preact'; // required imports
import { db } from '../db';
import Body from './Body';

export const UserIDCookieName = "continuewith_user_id"

export default async function HomePage(req: Request, res: Response) {
  const cookies = req.cookies
  const userID: string | undefined = cookies[UserIDCookieName]
  console.log('user id', userID)

  if (!userID) {
    return render(
      <UserNotFound />
    )
  }

  const user = await db.get(userID)
  if (!user) {
    res.clearCookie(UserIDCookieName)
    return render(
      <UserNotFound />
    )
  }

  return render(
    <Body>
      <h1>Welcome back, {user.Name}</h1>
    </Body>
  )
}


function UserNotFound() {
  return (
    <Body>
      <h1>Welcome unknown human!</h1>
      <p>You are not logged in, I'd love to know who you are, can you assist?</p>
      <form hx-post="/register" hx-target="html">
        <input name="name" type="text" placeholder="Name" />
        <button type="submit">Declare Humanity</button>
      </form>
    </Body>
  )
}
