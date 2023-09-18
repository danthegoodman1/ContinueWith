import { NextResponse } from 'next/server'
import { db } from './user'

// Give the user info to ContinueWith
export async function GET(request: Request) {
  const authHeader = request.headers.get("x-continuewith-auth")
  if (authHeader !== "supersecret") {
    return new Response("invalid auth", {status: 401})
  }

  // We'll pretend this was a JWT or session :)
  const userID = request.headers.get("user")

  try {
    const user = await db.get(userID || "")
    return NextResponse.json({
      UserID: user.ID
    })
  } catch (error) {
    return new Response("user not found", {status: 404})
  }
}
