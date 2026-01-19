import React from 'react';
import { useParams, useLocation } from 'wouter';
import { skipToken } from '@reduxjs/toolkit/query';
import { useDocumentBySlugQuery } from '@/store/api';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { MarkdownRenderer } from '@/components/MarkdownRenderer';
import { ArrowLeft, Edit, Calendar, Loader2 } from 'lucide-react';
import { format } from 'date-fns';

export default function DocumentView() {
  const params = useParams<{ slug: string }>();
  const [, navigate] = useLocation();
  const user = { id: 1, role: "admin" as const };

  const { data: document, isLoading, error } = useDocumentBySlugQuery(params.slug || skipToken);

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-slate-50 dark:bg-slate-950">
        <Loader2 className="h-8 w-8 animate-spin text-blue-600" />
      </div>
    );
  }

  if (error || !document) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-slate-50 dark:bg-slate-950">
        <div className="text-center">
          <h1 className="text-2xl font-bold mb-2">Document Not Found</h1>
          <p className="text-muted-foreground mb-4">
            The document you're looking for doesn't exist or you don't have access to it.
          </p>
          <Button onClick={() => navigate('/')}>
            <ArrowLeft className="mr-2 h-4 w-4" />
            Back to Home
          </Button>
        </div>
      </div>
    );
  }

  const canEdit = user?.role === 'admin' || user?.id === document.authorId;

  return (
    <div className="min-h-screen bg-slate-50 dark:bg-slate-950">
      {/* Header */}
      <header className="sticky top-0 z-50 bg-white dark:bg-slate-900 border-b border-slate-200 dark:border-slate-800">
        <div className="container flex items-center justify-between h-16">
          <Button variant="ghost" size="icon" onClick={() => navigate('/')}>
            <ArrowLeft className="h-5 w-5" />
          </Button>
          {canEdit && (
            <Button variant="outline" onClick={() => navigate(`/admin/edit/${document.id}`)}>
              <Edit className="mr-2 h-4 w-4" />
              Edit
            </Button>
          )}
        </div>
      </header>

      {/* Document Content */}
      <main className="container py-8">
        <article className="max-w-4xl mx-auto">
          {/* Document Header */}
          <header className="mb-8">
            <div className="flex items-center gap-2 mb-4">
              {document.category && (
                <Badge variant="secondary">{document.category}</Badge>
              )}
              {!document.isPublished && (
                <Badge variant="outline">Draft</Badge>
              )}
            </div>
            <h1 className="text-4xl font-bold tracking-tight mb-4">
              {document.title}
            </h1>
            {document.description && (
              <p className="text-xl text-muted-foreground mb-4">
                {document.description}
              </p>
            )}
            <div className="flex items-center gap-4 text-sm text-muted-foreground">
              <span className="flex items-center gap-1">
                <Calendar className="h-4 w-4" />
                {format(new Date(document.updatedAt), 'MMM d, yyyy')}
              </span>
            </div>
          </header>

          {/* Document Body */}
          <div className="bg-white dark:bg-slate-900 rounded-xl shadow-sm border border-slate-200 dark:border-slate-800 p-6 md:p-10">
            <MarkdownRenderer
              content={document.content}
              documentId={document.id}
              forms={document.forms}
              readOnly={false}
            />
          </div>
        </article>
      </main>
    </div>
  );
}
