import Link from "next/link";
import { Button } from "@/components/ui/button";
import { LandingNavbar } from "@/components/landing-navbar";
import {
  MessageSquare,
  Globe,
  Shield,
  BarChart3,
  ArrowRight,
  CheckCircle2,
  Phone,
  Mail,
  Cpu,
  Lock,
  Zap as ZapIcon,
} from "lucide-react";

export default function Home() {
  return (
    <div className="min-h-screen bg-black text-white selection:bg-white/20">
      <LandingNavbar />

      <section className="relative pt-32 pb-20 px-4 overflow-hidden">
        <div className="absolute inset-0 bg-[radial-gradient(ellipse_at_center,_var(--tw-gradient-stops))] from-zinc-800/20 via-black to-black" />
        <div className="absolute top-1/4 left-1/4 w-96 h-96 bg-primary/20 rounded-full blur-[120px]" />
        <div className="absolute bottom-1/4 right-1/4 w-96 h-96 bg-blue-500/10 rounded-full blur-[120px]" />

        <div className="container mx-auto relative max-w-5xl text-center">
          <div className="inline-flex items-center gap-2 rounded-full border border-white/10 bg-white/5 px-3 py-1 text-sm text-zinc-400 mb-8">
            <span className="flex h-1.5 w-1.5 rounded-full bg-green-500 animate-pulse"></span>
            Now with SMPP 3.4 Support
          </div>

          <h1 className="text-5xl md:text-7xl font-bold tracking-tight mb-6 bg-gradient-to-b from-white to-zinc-400 bg-clip-text text-transparent">
            Enterprise SMS Gateway
            <br />
            <span className="text-white">Built for Scale</span>
          </h1>

          <p className="text-lg md:text-xl text-zinc-400 mb-10 max-w-2xl mx-auto leading-relaxed">
            Connect to mobile networks worldwide with reliable, scalable
            infrastructure. Route messages through multiple carriers with
            intelligent failover.
          </p>

          <div className="flex flex-col sm:flex-row items-center justify-center gap-4">
            <Link href="/sign-up">
              <Button
                size="lg"
                className="bg-white text-black hover:bg-zinc-200 gap-2 w-full sm:w-auto h-12 px-8"
              >
                Start Sending
                <ArrowRight className="h-4 w-4" />
              </Button>
            </Link>
            <Link href="#features">
              <Button
                size="lg"
                variant="outline"
                className="border-zinc-700 text-zinc-300 hover:bg-white/5 hover:text-white w-full sm:w-auto h-12 px-8"
              >
                View Features
              </Button>
            </Link>
          </div>
        </div>
      </section>

      <section id="features" className="py-24 px-4 relative">
        <div className="container mx-auto">
          <div className="text-center mb-16">
            <h2 className="text-3xl md:text-4xl font-bold mb-4">
              Platform Features
            </h2>
            <p className="text-zinc-400 max-w-2xl mx-auto">
              Everything you need to build and scale your SMS business
            </p>
          </div>

          <div className="grid md:grid-cols-3 gap-6">
            <div className="group relative rounded-2xl border border-white/10 bg-zinc-900/50 p-8 hover:border-white/20 transition-all">
              <div className="absolute inset-0 bg-gradient-to-b from-white/5 to-transparent opacity-0 group-hover:opacity-100 transition-opacity rounded-2xl" />
              <div className="relative">
                <div className="w-12 h-12 rounded-xl bg-white/5 flex items-center justify-center mb-4 border border-white/10">
                  <Globe className="h-6 w-6 text-blue-400" />
                </div>
                <h3 className="text-xl font-semibold mb-2">
                  Global Connectivity
                </h3>
                <p className="text-zinc-400 text-sm leading-relaxed">
                  Direct connections to mobile networks across Africa, Asia, and
                  the Middle East with 99.9% deliverability.
                </p>
              </div>
            </div>

            <div className="group relative rounded-2xl border border-white/10 bg-zinc-900/50 p-8 hover:border-white/20 transition-all">
              <div className="absolute inset-0 bg-gradient-to-b from-white/5 to-transparent opacity-0 group-hover:opacity-100 transition-opacity rounded-2xl" />
              <div className="relative">
                <div className="w-12 h-12 rounded-xl bg-white/5 flex items-center justify-center mb-4 border border-white/10">
                  <ZapIcon className="h-6 w-6 text-amber-400" />
                </div>
                <h3 className="text-xl font-semibold mb-2">
                  Real-Time Processing
                </h3>
                <p className="text-zinc-400 text-sm leading-relaxed">
                  Sub-second message processing with instant delivery receipts
                  and webhook notifications.
                </p>
              </div>
            </div>

            <div className="group relative rounded-2xl border border-white/10 bg-zinc-900/50 p-8 hover:border-white/20 transition-all">
              <div className="absolute inset-0 bg-gradient-to-b from-white/5 to-transparent opacity-0 group-hover:opacity-100 transition-opacity rounded-2xl" />
              <div className="relative">
                <div className="w-12 h-12 rounded-xl bg-white/5 flex items-center justify-center mb-4 border border-white/10">
                  <Cpu className="h-6 w-6 text-emerald-400" />
                </div>
                <h3 className="text-xl font-semibold mb-2">Smart Routing</h3>
                <p className="text-zinc-400 text-sm leading-relaxed">
                  AI-powered routing that automatically selects the best carrier
                  based on cost, latency, and delivery rates.
                </p>
              </div>
            </div>

            <div className="group relative rounded-2xl border border-white/10 bg-zinc-900/50 p-8 hover:border-white/20 transition-all">
              <div className="absolute inset-0 bg-gradient-to-b from-white/5 to-transparent opacity-0 group-hover:opacity-100 transition-opacity rounded-2xl" />
              <div className="relative">
                <div className="w-12 h-12 rounded-xl bg-white/5 flex items-center justify-center mb-4 border border-white/10">
                  <Lock className="h-6 w-6 text-purple-400" />
                </div>
                <h3 className="text-xl font-semibold mb-2">OTP & 2FA</h3>
                <p className="text-zinc-400 text-sm leading-relaxed">
                  Optimized sender IDs and templates for secure authentication
                  and verification use cases.
                </p>
              </div>
            </div>

            <div className="group relative rounded-2xl border border-white/10 bg-zinc-900/50 p-8 hover:border-white/20 transition-all">
              <div className="absolute inset-0 bg-gradient-to-b from-white/5 to-transparent opacity-0 group-hover:opacity-100 transition-opacity rounded-2xl" />
              <div className="relative">
                <div className="w-12 h-12 rounded-xl bg-white/5 flex items-center justify-center mb-4 border border-white/10">
                  <BarChart3 className="h-6 w-6 text-rose-400" />
                </div>
                <h3 className="text-xl font-semibold mb-2">
                  Analytics Dashboard
                </h3>
                <p className="text-zinc-400 text-sm leading-relaxed">
                  Real-time insights into message delivery, costs, and
                  performance with exportable reports.
                </p>
              </div>
            </div>

            <div className="group relative rounded-2xl border border-white/10 bg-zinc-900/50 p-8 hover:border-white/20 transition-all">
              <div className="absolute inset-0 bg-gradient-to-b from-white/5 to-transparent opacity-0 group-hover:opacity-100 transition-opacity rounded-2xl" />
              <div className="relative">
                <div className="w-12 h-12 rounded-xl bg-white/5 flex items-center justify-center mb-4 border border-white/10">
                  <Shield className="h-6 w-6 text-cyan-400" />
                </div>
                <h3 className="text-xl font-semibold mb-2">99.9% Uptime</h3>
                <p className="text-zinc-400 text-sm leading-relaxed">
                  Enterprise-grade infrastructure with redundant connections and
                  automatic failover.
                </p>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section className="py-24 px-4 relative overflow-hidden">
        <div className="absolute inset-0 bg-gradient-to-b from-zinc-900/50 to-black" />
        <div className="container mx-auto relative">
          <div className="grid md:grid-cols-2 gap-12 items-center">
            <div>
              <h2 className="text-3xl md:text-4xl font-bold mb-4">
                Developer-First API
              </h2>
              <p className="text-zinc-400 text-lg mb-8 leading-relaxed">
                Built for developers, by developers. Integrate in minutes with
                our RESTful API and comprehensive SDKs.
              </p>
              <div className="space-y-3">
                <div className="flex items-center gap-3 text-zinc-300">
                  <CheckCircle2 className="h-5 w-5 text-green-500" />
                  <span>RESTful API with JSON responses</span>
                </div>
                <div className="flex items-center gap-3 text-zinc-300">
                  <CheckCircle2 className="h-5 w-5 text-green-500" />
                  <span>SMPP 3.4 protocol support</span>
                </div>
                <div className="flex items-center gap-3 text-zinc-300">
                  <CheckCircle2 className="h-5 w-5 text-green-500" />
                  <span>Webhook delivery receipts</span>
                </div>
                <div className="flex items-center gap-3 text-zinc-300">
                  <CheckCircle2 className="h-5 w-5 text-green-500" />
                  <span>Client SDKs for Node, Python, Go</span>
                </div>
              </div>
            </div>
            <div className="relative">
              <div className="absolute inset-0 bg-gradient-to-r from-blue-500/20 to-purple-500/20 blur-[100px]" />
              <div className="relative rounded-xl border border-white/10 bg-zinc-900/80 backdrop-blur overflow-hidden">
                <div className="flex items-center gap-2 px-4 py-3 border-b border-white/10 bg-white/5">
                  <div className="w-3 h-3 rounded-full bg-red-500" />
                  <div className="w-3 h-3 rounded-full bg-yellow-500" />
                  <div className="w-3 h-3 rounded-full bg-green-500" />
                  <span className="ml-2 text-xs text-zinc-500">bash</span>
                </div>
                <div className="p-6 font-mono text-sm">
                  <div className="text-zinc-400 mb-2">
                    # Send an SMS message
                  </div>
                  <div className="text-green-400">
                    curl -X POST https://api.aegisbox.com/v1/send \
                  </div>
                  <div className="text-zinc-300 ml-4">
                    -H "Authorization: Bearer YOUR_API_KEY" \
                  </div>
                  <div className="text-zinc-300 ml-4">
                    -d &#39;&#123;&#34;to&#34;: &#34;+2348012345678&#34;,
                    &#34;message&#34;: &#34;Hello!&#34;&#39;
                  </div>
                  <div className="text-zinc-500 mt-4">{`{`}</div>
                  <div className="text-zinc-300 ml-4">"status": "queued",</div>
                  <div className="text-zinc-300 ml-4">
                    "message_id": "msg_abc123",
                  </div>
                  <div className="text-zinc-300 ml-4">
                    "to": "+2348012345678"
                  </div>
                  <div className="text-zinc-500">{`}`}</div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section id="networks" className="py-24 px-4">
        <div className="container mx-auto">
          <div className="text-center mb-16">
            <h2 className="text-3xl md:text-4xl font-bold mb-4">
              Global Network
            </h2>
            <p className="text-zinc-400 max-w-2xl mx-auto">
              Direct connections to major mobile operators worldwide
            </p>
          </div>

          <div className="grid md:grid-cols-4 gap-4 max-w-4xl mx-auto">
            <div className="rounded-xl border border-white/10 bg-zinc-900/50 p-6 text-center hover:border-white/20 transition-all">
              <h4 className="font-semibold text-white mb-2">Nigeria</h4>
              <p className="text-sm text-zinc-400">MTN, Glo, Airtel, 9mobile</p>
            </div>
            <div className="rounded-xl border border-white/10 bg-zinc-900/50 p-6 text-center hover:border-white/20 transition-all">
              <h4 className="font-semibold text-white mb-2">Ghana</h4>
              <p className="text-sm text-zinc-400">MTN, Vodafone, AirtelTigo</p>
            </div>
            <div className="rounded-xl border border-white/10 bg-zinc-900/50 p-6 text-center hover:border-white/20 transition-all">
              <h4 className="font-semibold text-white mb-2">Egypt</h4>
              <p className="text-sm text-zinc-400">
                Vodafone, Orange, Etisalat, WE
              </p>
            </div>
            <div className="rounded-xl border border-white/10 bg-zinc-900/50 p-6 text-center hover:border-white/20 transition-all">
              <h4 className="font-semibold text-white mb-2">Coming Soon</h4>
              <p className="text-sm text-zinc-400">
                Kenya, South Africa, Uganda
              </p>
            </div>
          </div>
        </div>
      </section>

      <section className="py-24 px-4 relative">
        <div className="absolute inset-0 bg-gradient-to-t from-zinc-900/50 to-transparent" />
        <div className="container mx-auto relative text-center max-w-3xl">
          <h2 className="text-3xl md:text-4xl font-bold mb-4">
            Ready to Scale?
          </h2>
          <p className="text-zinc-400 text-lg mb-8">
            Create your account in minutes and start sending SMS worldwide
          </p>
          <div className="flex flex-col sm:flex-row items-center justify-center gap-4">
            <Link href="/sign-up">
              <Button
                size="lg"
                className="bg-white text-black hover:bg-zinc-200 gap-2 h-12 px-8"
              >
                Create Free Account
                <ArrowRight className="h-4 w-4" />
              </Button>
            </Link>
          </div>
        </div>
      </section>

      <footer className="border-t border-white/10 py-12 px-4 bg-black">
        <div className="container mx-auto">
          <div className="grid md:grid-cols-4 gap-8 mb-8">
            <div>
              <div className="flex items-center gap-2 mb-4">
                <div className="flex h-6 w-6 items-center justify-center rounded-md bg-white text-black">
                  <MessageSquare className="h-4 w-4" />
                </div>
                <span className="font-bold">Aegisbox</span>
              </div>
              <p className="text-sm text-zinc-500">
                Enterprise SMS gateway platform for businesses worldwide.
              </p>
            </div>
            <div>
              <h4 className="font-semibold mb-3">Product</h4>
              <ul className="space-y-2 text-sm text-zinc-500">
                <li>
                  <Link
                    href="#"
                    className="hover:text-zinc-300 transition-colors"
                  >
                    Features
                  </Link>
                </li>
                <li>
                  <Link
                    href="#"
                    className="hover:text-zinc-300 transition-colors"
                  >
                    Pricing
                  </Link>
                </li>
                <li>
                  <Link
                    href="#"
                    className="hover:text-zinc-300 transition-colors"
                  >
                    API Docs
                  </Link>
                </li>
              </ul>
            </div>
            <div>
              <h4 className="font-semibold mb-3">Company</h4>
              <ul className="space-y-2 text-sm text-zinc-500">
                <li>
                  <Link
                    href="#"
                    className="hover:text-zinc-300 transition-colors"
                  >
                    About
                  </Link>
                </li>
                <li>
                  <Link
                    href="#"
                    className="hover:text-zinc-300 transition-colors"
                  >
                    Blog
                  </Link>
                </li>
                <li>
                  <Link
                    href="#"
                    className="hover:text-zinc-300 transition-colors"
                  >
                    Contact
                  </Link>
                </li>
              </ul>
            </div>
            <div>
              <h4 className="font-semibold mb-3">Contact</h4>
              <ul className="space-y-2 text-sm text-zinc-500">
                <li className="flex items-center gap-2">
                  <Mail className="h-4 w-4" />
                  info@thrillee.tech
                </li>
                <li className="flex items-center gap-2">
                  <Phone className="h-4 w-4" />
                  +234 815 439 6918
                </li>
              </ul>
            </div>
          </div>
          <div className="pt-6 border-t border-white/10 flex flex-col md:flex-row items-center justify-between gap-4 text-sm text-zinc-600">
            <p>
              ©{new Date().getFullYear()} Thrillee Technologies. All rights
              reserved.
            </p>
            <div className="flex items-center gap-4">
              <Link href="#" className="hover:text-zinc-400 transition-colors">
                Privacy
              </Link>
              <Link href="#" className="hover:text-zinc-400 transition-colors">
                Terms
              </Link>
            </div>
          </div>
        </div>
      </footer>
    </div>
  );
}
