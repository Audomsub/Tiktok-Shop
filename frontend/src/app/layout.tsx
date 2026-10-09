import type { Metadata } from "next";
import { Inter } from "next/font/google";
import "./globals.css";
import { Navbar } from "@/components/navbar";
import { AuthProvider } from "@/context/auth-context";
import { FavoritesProvider } from "@/context/favorites-context";
import { AuthModal } from "@/components/auth-modal";

const inter = Inter({ subsets: ["latin"] });

export const metadata: Metadata = {
  title: "TikTok Affiliate Winning Product Analytics",
  description: "Real-time algorithmic scoring and winning product radar for TikTok Shop creators and affiliate marketers.",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" className="dark">
      <body className={`${inter.className} min-h-screen bg-background text-foreground antialiased flex flex-col`}>
        <AuthProvider>
          <FavoritesProvider>
            <Navbar />
            <main className="flex-1">{children}</main>
            <AuthModal />
            <footer className="border-t border-border/40 py-6 text-center text-xs text-muted-foreground bg-background">
              <p>© 2026 AffiliatePulse. Powered by Go Clean Architecture & Supabase Engine.</p>
            </footer>
          </FavoritesProvider>
        </AuthProvider>
      </body>
    </html>
  );
}
