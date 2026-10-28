import { SignUp } from '@clerk/nextjs'

export default function SignUpPage() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-black p-4">
      <div className="w-full max-w-md">
        <SignUp 
          routing="path" 
          path="/sign-up"
          signInUrl="/sign-in"
          forceRedirectUrl="/auth-check"
        />
      </div>
    </div>
  );
}
