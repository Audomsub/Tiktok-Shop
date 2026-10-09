import { createBrowserClient } from "@supabase/ssr";

let supabaseBrowserClient: ReturnType<typeof createBrowserClient> | null = null;

export function getSupabaseBrowserClient() {
  if (supabaseBrowserClient) {
    return supabaseBrowserClient;
  }

  const supabaseUrl = process.env.NEXT_PUBLIC_SUPABASE_URL || "https://evyltfcspqzctgdehxff.supabase.co";
  const supabaseAnonKey =
    process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY ||
    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6ImV2eWx0ZmNzcHF6Y3RnZGVoeGZmIiwicm9sZSI6ImFub24iLCJpYXQiOjE3OTE0NDA0NzUsImV4cCI6MjEwNzAxNjQ3NX0.KP3wwTCfli28bboe9w8-kHPBeTH74WuwUpwH5rklxXQ";

  supabaseBrowserClient = createBrowserClient(supabaseUrl, supabaseAnonKey);
  return supabaseBrowserClient;
}
