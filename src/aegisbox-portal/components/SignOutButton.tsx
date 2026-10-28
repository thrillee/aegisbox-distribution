'use client';

import React from "react"

import Link from 'next/link';
import { Button } from '@/components/ui/button';

// Placeholder SignOutButton - just navigates back to home
// Replace with Clerk SignOutButton when authentication is configured
export default function SignOutButton({ children }: { children?: React.ReactNode }) {
  return (
    <Link href="/">
      {children || <Button variant="outline" className="bg-transparent">Sign Out</Button>}
    </Link>
  );
}
