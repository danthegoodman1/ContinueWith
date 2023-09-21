export const runtime = 'nodejs'

import { randomUUID } from "crypto"
import { NextResponse } from "next/server"
import { Level } from "level"
export const db = new Level<string, User>("./db", { valueEncoding: "json" })

interface CreateUserRequest {
  Name: string
}

interface User {
  Name: string
  ID: string
}

interface ContinueWithTokenResponse {
  UserID: string
  Created: number
  Expires: number
  Scopes: string[]
}

// Get a user
export async function GET(request: Request) {
  const authHeader = request.headers.get("Authorization")?.split("earer ")[1]
  if (!authHeader) {
    return new Response("missing auth header", {status: 404})
  }

  if (authHeader.startsWith("a_")) {
    // This is from the client
    // Get the user from ContinueWith
    const res = await fetch(`http://localhost:8880/admin/access_token/${encodeURIComponent(authHeader)}`)
    const accessTokenInfo: ContinueWithTokenResponse = await res.json()
    const user = await db.get(accessTokenInfo.UserID)!
    // This is a protected API endpoint
    return NextResponse.json(user)
  }

  // This is from our site (provider)
  try {
    const user = await db.get(authHeader) // it's just the UserID to keep it simple
    return NextResponse.json(user)
  } catch (error) {
    console.error("Error getting user")
    console.error(error)
    return new Response((error as Error).message, {status: 500})
  }
}

// Create a user
export async function POST(request: Request) {
  const body: CreateUserRequest = await request.json()
  const user = {
    Name: body.Name,
    ID: randomUUID()
  }

  await db.put(user.ID, user)

  return NextResponse.json(user)
}
