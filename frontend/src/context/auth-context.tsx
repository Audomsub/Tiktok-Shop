"use client";

import React, { createContext, useContext, useEffect, useState } from "react";
import { User, Session, AuthError, AuthChangeEvent } from "@supabase/supabase-js";
import { getSupabaseBrowserClient } from "@/lib/supabase/client";

interface AuthContextType {
  user: User | null;
  session: Session | null;
  accessToken: string | null;
  isLoading: boolean;
  isAuthModalOpen: boolean;
  authModalMessage: string;
  openAuthModal: (message?: string) => void;
  closeAuthModal: () => void;
  signInWithEmail: (email: string, password: string) => Promise<{ error: AuthError | null }>;
  signUpWithEmail: (email: string, password: string) => Promise<{ error: AuthError | null; user: User | null }>;
  signOut: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [session, setSession] = useState<Session | null>(null);
  const [accessToken, setAccessToken] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isAuthModalOpen, setIsAuthModalOpen] = useState(false);
  const [authModalMessage, setAuthModalMessage] = useState("");

  const supabase = getSupabaseBrowserClient();

  useEffect(() => {
    let mounted = true;

    async function initAuth() {
      try {
        const { data, error } = await supabase.auth.getSession();
        if (error) {
          console.error("Error retrieving Supabase session:", error);
        }
        if (mounted) {
          setSession(data.session);
          setUser(data.session?.user ?? null);
          setAccessToken(data.session?.access_token ?? null);
          setIsLoading(false);
        }
      } catch (err) {
        console.error("Supabase auth init exception:", err);
        if (mounted) setIsLoading(false);
      }
    }

    initAuth();

    const { data: authListener } = supabase.auth.onAuthStateChange(
      (_event: AuthChangeEvent, currentSession: Session | null) => {
      if (!mounted) return;
      setSession(currentSession);
      setUser(currentSession?.user ?? null);
      setAccessToken(currentSession?.access_token ?? null);
      setIsLoading(false);
    });

    return () => {
      mounted = false;
      authListener.subscription.unsubscribe();
    };
  }, [supabase]);

  const openAuthModal = (message?: string) => {
    setAuthModalMessage(
      message || "Sign in or create an account to unlock creator bookmarks and personalized radar."
    );
    setIsAuthModalOpen(true);
  };

  const closeAuthModal = () => {
    setIsAuthModalOpen(false);
    setAuthModalMessage("");
  };

  const signInWithEmail = async (email: string, password: string) => {
    const { data, error } = await supabase.auth.signInWithPassword({
      email,
      password,
    });
    if (!error && data.session) {
      setSession(data.session);
      setUser(data.session.user);
      setAccessToken(data.session.access_token);
      closeAuthModal();
    }
    return { error };
  };

  const signUpWithEmail = async (email: string, password: string) => {
    const { data, error } = await supabase.auth.signUp({
      email,
      password,
    });
    if (!error && data.session) {
      setSession(data.session);
      setUser(data.session.user);
      setAccessToken(data.session.access_token);
      closeAuthModal();
    }
    return { error, user: data.user };
  };

  const signOut = async () => {
    try {
      await supabase.auth.signOut();
    } catch (err) {
      console.error("Sign out error:", err);
    } finally {
      setUser(null);
      setSession(null);
      setAccessToken(null);
    }
  };

  return (
    <AuthContext.Provider
      value={{
        user,
        session,
        accessToken,
        isLoading,
        isAuthModalOpen,
        authModalMessage,
        openAuthModal,
        closeAuthModal,
        signInWithEmail,
        signUpWithEmail,
        signOut,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
}
