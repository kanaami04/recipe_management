import { useState } from 'react'
import { useNavigate } from 'react-router-dom'

import {
  Card,
  CardAction,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { login } from '@/shared/auth/authClient'

import { MessageAlertDialog } from './MessageAlertDialog.tsx'
import { Button } from './ui/button.tsx'

export function UserLoginFrom() {
  const navigate = useNavigate()
  const [userName, setUserName] = useState('')
  const [userPassword, setUserPassword] = useState('')
  const [isErrorOpen, setIsErrorOpen] = useState(false)

  const handleLogin = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()

    try {
      // access はメモリ保持、refresh は httpOnly Cookie で発行される (ADR-0004)。
      await login(userName, userPassword)
      navigate('/top')
    } catch (error) {
      console.error(error)
      setIsErrorOpen(true)
    }
  }

  return (
    <>
      <Card className="w-full max-w-sm">
        <CardHeader>
          <CardTitle>Login to your account</CardTitle>
          <CardDescription>Enter your email below to login to your account</CardDescription>
          <CardAction>
            <Button variant="link">Sign Up</Button>
          </CardAction>
        </CardHeader>
        <CardContent>
          <div className="flex flex-col gap-6">
            <div className="grid gap-2">
              <Label htmlFor="username">Username</Label>
              <Input
                id="username"
                type="text"
                value={userName}
                onChange={(e) => setUserName(e.target.value)}
                required
              />
            </div>
            <div className="grid gap-2">
              <div className="flex items-center">
                <Label htmlFor="password">Password</Label>
                <a
                  href="#"
                  className="ml-auto inline-block text-sm underline-offset-4 hover:underline"
                >
                  Forgot your password?
                </a>
              </div>
              <Input
                id="password"
                type="password"
                value={userPassword}
                onChange={(e) => setUserPassword(e.target.value)}
                required
              />
            </div>
          </div>
        </CardContent>
        <CardFooter className="flex-col gap-2">
          <form className="w-full" onSubmit={handleLogin}>
            <Button type="submit" className="w-full">
              Login
            </Button>
          </form>
        </CardFooter>
      </Card>

      <MessageAlertDialog
        title="認証に失敗しました"
        description={`ユーザー名またはパスワードが間違っています。\nもう一度入力してください。`}
        open={isErrorOpen}
        onOpenChange={() => setIsErrorOpen(false)}
      />
    </>
  )
}
