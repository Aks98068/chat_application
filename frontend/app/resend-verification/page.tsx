
"use client";

import { FormEvent, useState } from "react";
import Link from "next/link";
import { CheckCircle2, Loader2, Mail, XCircle } from "lucide-react";

import { apiRequest } from "@/lib/api";

type ResendVerificationResponse = {
  message?: string;
};

export default function ResendVerificationPage() {
  const [email, setEmail] = useState("");
  const [status, setStatus] = useState<
    "idle" | "loading" | "success" | "error"
  >("idle");
  const [message, setMessage] = useState("");

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    setStatus("loading");
    setMessage("");

    try {
      const response =
        await apiRequest<ResendVerificationResponse>(
          "/api/v1/auth/resend-verification",
          {
            method: "POST",
            body: JSON.stringify({
              email: email.trim(),
            }),
          }
        );

      setStatus("success");
      setMessage(
        response.message ||
          "If an account exists with this email, a new verification email has been sent."
      );
    } catch (error) {
      setStatus("error");
      setMessage(
        error instanceof Error
          ? error.message
          : "Unable to resend the verification email."
      );
    }
  }

  return (
    <main className="min-h-screen bg-slate-50 flex items-center justify-center px-4 py-12">
      <div className="w-full max-w-md">
        <div className="rounded-2xl border border-slate-200 bg-white p-8 shadow-sm">
          {/* Header */}
          <div className="mb-8 text-center">
            <div className="mx-auto mb-5 flex h-14 w-14 items-center justify-center rounded-full bg-indigo-50">
              <Mail className="h-7 w-7 text-indigo-600" />
            </div>

            <h1 className="text-2xl font-semibold text-slate-900">
              Resend verification email
            </h1>

            <p className="mt-2 text-sm leading-6 text-slate-500">
              Enter your email address and we&apos;ll send you a new
              verification link.
            </p>
          </div>

          {/* Success */}
          {status === "success" && (
            <div className="mb-6 rounded-xl border border-green-200 bg-green-50 p-4">
              <div className="flex items-start gap-3">
                <CheckCircle2 className="mt-0.5 h-5 w-5 shrink-0 text-green-600" />

                <div>
                  <p className="text-sm font-medium text-green-800">
                    Verification email sent
                  </p>

                  <p className="mt-1 text-sm leading-5 text-green-700">
                    {message}
                  </p>
                </div>
              </div>
            </div>
          )}

          {/* Error */}
          {status === "error" && (
            <div className="mb-6 rounded-xl border border-red-200 bg-red-50 p-4">
              <div className="flex items-start gap-3">
                <XCircle className="mt-0.5 h-5 w-5 shrink-0 text-red-600" />

                <div>
                  <p className="text-sm font-medium text-red-800">
                    Unable to resend email
                  </p>

                  <p className="mt-1 text-sm leading-5 text-red-700">
                    {message}
                  </p>
                </div>
              </div>
            </div>
          )}

          {/* Form */}
          <form onSubmit={handleSubmit} className="space-y-5">
            <div>
              <label
                htmlFor="email"
                className="mb-2 block text-sm font-medium text-slate-700"
              >
                Email address
              </label>

              <input
                id="email"
                name="email"
                type="email"
                autoComplete="email"
                required
                value={email}
                onChange={(event) => setEmail(event.target.value)}
                placeholder="you@example.com"
                disabled={status === "loading"}
                className="w-full rounded-xl border border-slate-300 bg-white px-4 py-3 text-sm text-slate-900 outline-none transition placeholder:text-slate-400 focus:border-indigo-500 focus:ring-2 focus:ring-indigo-100 disabled:cursor-not-allowed disabled:bg-slate-50"
              />
            </div>

            <button
              type="submit"
              disabled={status === "loading"}
              className="flex w-full items-center justify-center gap-2 rounded-xl bg-indigo-600 px-4 py-3 text-sm font-semibold text-white transition hover:bg-indigo-700 disabled:cursor-not-allowed disabled:opacity-60"
            >
              {status === "loading" ? (
                <>
                  <Loader2 className="h-4 w-4 animate-spin" />
                  Sending...
                </>
              ) : (
                "Resend verification email"
              )}
            </button>
          </form>

          {/* Links */}
          <div className="mt-6 flex flex-col items-center gap-3 text-sm">
            <Link
              href="/verify-email"
              className="font-medium text-indigo-600 hover:text-indigo-700"
            >
              Back to email verification
            </Link>

            <Link
              href="/login"
              className="text-slate-500 hover:text-slate-700"
            >
              Back to login
            </Link>
          </div>
        </div>
      </div>
    </main>
  );
}

