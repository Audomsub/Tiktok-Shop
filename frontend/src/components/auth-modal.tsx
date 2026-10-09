"use client";

import React, { useState } from "react";
import { useAuth } from "@/context/auth-context";
import { X, Lock, Mail, Loader2, Sparkles, AlertCircle, CheckCircle } from "lucide-react";

export function AuthModal() {
  const { isAuthModalOpen, authModalMessage, closeAuthModal, signInWithEmail, signUpWithEmail } = useAuth();
  const [isSignUp, setIsSignUp] = useState(false);
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [loading, setLoading] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);

  if (!isAuthModalOpen) return null;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setErrorMessage(null);
    setSuccessMessage(null);
    setLoading(true);

    try {
      if (isSignUp) {
        const { error, user } = await signUpWithEmail(email, password);
        if (error) {
          setErrorMessage(error.message);
        } else if (user && !user.confirmed_at && user.identities?.length) {
          setSuccessMessage("Account created! Please check your email to confirm your registration.");
        }
      } else {
        const { error } = await signInWithEmail(email, password);
        if (error) {
          setErrorMessage(error.message);
        }
      }
    } catch (err: unknown) {
      const errObj = err as Error;
      setErrorMessage(errObj.message || "An unexpected error occurred.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-background/80 backdrop-blur-sm animate-in fade-in duration-200">
      <div className="relative w-full max-w-md overflow-hidden rounded-2xl bg-card border border-border shadow-2xl p-6 sm:p-8">
        {/* Close Button */}
        <button
          type="button"
          onClick={closeAuthModal}
          className="absolute right-4 top-4 rounded-xl p-1.5 text-muted-foreground hover:bg-secondary hover:text-foreground transition-colors"
        >
          <X className="h-5 w-5" />
        </button>

        {/* Modal Header */}
        <div className="flex flex-col items-center text-center mb-6">
          <div className="flex h-12 w-12 items-center justify-center rounded-2xl bg-gradient-to-tr from-[#fe2c55] to-[#25f4ee] p-0.5 shadow-lg shadow-[#fe2c55]/20 mb-3">
            <div className="flex h-full w-full items-center justify-center rounded-[14px] bg-background">
              <Sparkles className="h-6 w-6 text-[#25f4ee]" />
            </div>
          </div>
          <h2 className="text-xl font-bold tracking-tight text-foreground">
            {isSignUp ? "Create Creator Account" : "Welcome Back"}
          </h2>
          <p className="mt-1 text-xs text-muted-foreground max-w-xs">
            {authModalMessage || "Sign in to access your bookmarked winning products and custom radar."}
          </p>
        </div>

        {/* Alert Messages */}
        {errorMessage && (
          <div className="mb-4 flex items-start gap-2.5 rounded-xl bg-destructive/15 border border-destructive/30 p-3 text-xs text-destructive">
            <AlertCircle className="h-4 w-4 flex-shrink-0 mt-0.5" />
            <p className="font-medium">{errorMessage}</p>
          </div>
        )}

        {successMessage && (
          <div className="mb-4 flex items-start gap-2.5 rounded-xl bg-emerald-500/15 border border-emerald-500/30 p-3 text-xs text-emerald-400">
            <CheckCircle className="h-4 w-4 flex-shrink-0 mt-0.5" />
            <p className="font-medium">{successMessage}</p>
          </div>
        )}

        {/* Form */}
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-1.5">
            <label className="text-xs font-medium text-foreground" htmlFor="auth-email">
              Email Address
            </label>
            <div className="relative">
              <Mail className="absolute left-3.5 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
              <input
                id="auth-email"
                type="email"
                required
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="creator@tiktok.com"
                className="w-full rounded-xl bg-secondary/60 border border-border/80 pl-10 pr-4 py-2.5 text-sm text-foreground placeholder:text-muted-foreground/60 focus:outline-none focus:ring-2 focus:ring-[#fe2c55]/50 focus:border-[#fe2c55]"
              />
            </div>
          </div>

          <div className="space-y-1.5">
            <label className="text-xs font-medium text-foreground" htmlFor="auth-password">
              Password
            </label>
            <div className="relative">
              <Lock className="absolute left-3.5 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
              <input
                id="auth-password"
                type="password"
                required
                minLength={6}
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="••••••••"
                className="w-full rounded-xl bg-secondary/60 border border-border/80 pl-10 pr-4 py-2.5 text-sm text-foreground placeholder:text-muted-foreground/60 focus:outline-none focus:ring-2 focus:ring-[#fe2c55]/50 focus:border-[#fe2c55]"
              />
            </div>
          </div>

          <button
            type="submit"
            disabled={loading}
            className="w-full mt-2 flex items-center justify-center gap-2 rounded-xl bg-gradient-to-r from-[#fe2c55] to-[#fe2c55]/90 hover:opacity-95 text-white py-2.5 text-sm font-semibold shadow-lg shadow-[#fe2c55]/25 transition-all disabled:opacity-50"
          >
            {loading && <Loader2 className="h-4 w-4 animate-spin" />}
            <span>{isSignUp ? "Sign Up" : "Sign In"}</span>
          </button>
        </form>

        {/* Toggle between Sign In / Sign Up */}
        <div className="mt-5 text-center text-xs text-muted-foreground">
          {isSignUp ? (
            <p>
              Already have an account?{" "}
              <button
                type="button"
                onClick={() => {
                  setIsSignUp(false);
                  setErrorMessage(null);
                  setSuccessMessage(null);
                }}
                className="font-semibold text-[#25f4ee] hover:underline"
              >
                Sign In
              </button>
            </p>
          ) : (
            <p>
              Don&apos;t have an account?{" "}
              <button
                type="button"
                onClick={() => {
                  setIsSignUp(true);
                  setErrorMessage(null);
                  setSuccessMessage(null);
                }}
                className="font-semibold text-[#25f4ee] hover:underline"
              >
                Create one now
              </button>
            </p>
          )}
        </div>
      </div>
    </div>
  );
}
