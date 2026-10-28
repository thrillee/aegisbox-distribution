import { redirect } from "next/navigation";
import { auth } from "@clerk/nextjs/server";
import { DashboardSidebar } from "@/components/dashboard/sidebar";

export default async function AdminLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const { userId, orgId, sessionClaims } = await auth();

  const adminOrgId = process.env.CLERK_ADMIN_ORG_ID;
  const isAdmin =
    orgId === adminOrgId || (sessionClaims?.metadata as any)?.role === "admin";

  if (!isAdmin) {
    redirect("/sp/dashboard");
  }

  return (
    <div className="min-h-screen bg-black">
      <DashboardSidebar variant="admin" />
      <main className="lg:ml-64 min-h-screen">{children}</main>
    </div>
  );
}
