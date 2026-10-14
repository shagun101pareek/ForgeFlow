import type { ReactNode } from "react";

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Wordmark } from "@/components/wordmark";

export function AuthFrame({
  title,
  description,
  children,
  footer,
}: {
  title: string;
  description: string;
  children: ReactNode;
  footer: ReactNode;
}) {
  return (
    <main className="flex flex-1 flex-col px-6 py-8">
      <Wordmark />
      <div className="flex flex-1 items-center justify-center py-10">
        <Card className="w-full max-w-md">
          <CardHeader>
            <CardTitle>
              <h1 className="font-heading text-3xl font-medium">{title}</h1>
            </CardTitle>
            <CardDescription>{description}</CardDescription>
          </CardHeader>
          <CardContent>
            {children}
            <div className="mt-4 text-sm text-muted-foreground">{footer}</div>
          </CardContent>
        </Card>
      </div>
    </main>
  );
}
