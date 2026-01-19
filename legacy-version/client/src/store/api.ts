import { createApi, fetchBaseQuery } from "@reduxjs/toolkit/query/react";

export type Document = {
  id: number;
  title: string;
  slug: string;
  content: string;
  description: string | null;
  category: string | null;
  isPublished: boolean;
  authorId: number;
  createdAt: string;
  updatedAt: string;
};

export type QuizForm = {
  formId: string;
  definition: unknown;
};

export type DocumentWithForms = Document & {
  forms: QuizForm[];
};

export type QuizSubmission = {
  id: number;
  userId: number;
  documentId: number;
  formId: string;
  responses: Record<string, unknown>;
  score: number | null;
  maxScore: number | null;
  submittedAt: string;
};

export type DocumentAnalytics = {
  totalSubmissions: number;
  averageScore: number;
  highestScore: number;
  lowestScore: number;
};

export type SubmissionWithUser = {
  submission: QuizSubmission;
  userName: string | null;
};

export type SubmissionWithDocument = {
  submission: QuizSubmission;
  documentTitle: string;
  documentSlug: string;
};

export type SubmissionDetail = {
  submission: QuizSubmission;
  documentTitle: string;
  documentSlug: string;
  formDefinition: unknown;
};

export type CreateDocumentRequest = {
  title: string;
  content: string;
  description?: string;
  category?: string;
  isPublished: boolean;
};

export type CreateDocumentResponse = {
  id: number;
  slug: string;
};

export type UpdateDocumentRequest = {
  id: number;
  title?: string;
  content?: string;
  description?: string;
  category?: string;
  isPublished?: boolean;
};

export type SuccessResponse = {
  success: true;
};

export type SubmitQuizRequest = {
  documentId: number;
  formId: string;
  responses: Record<string, unknown>;
};

export type SubmitQuizResponse = {
  id: number;
  score: number | null;
  maxScore: number | null;
};

export type SubmitQuizBatchRequest = {
  documentId: number;
  submissions: Array<{ formId: string; responses: Record<string, unknown> }>;
};

export type SubmitQuizBatchResponse = {
  results: Array<{ formId: string; score: number; maxScore: number }>;
};

export const api = createApi({
  reducerPath: "api",
  baseQuery: fetchBaseQuery({ baseUrl: "/api", credentials: "include" }),
  tagTypes: [
    "Documents",
    "MyDocuments",
    "Document",
    "MySubmissions",
    "DocAnalytics",
    "DocSubmissions",
    "Submission",
  ],
  endpoints: build => ({
    listDocuments: build.query<Document[], void>({
      query: () => ({ url: "documents", params: { scope: "all" } }),
      providesTags: ["Documents"],
    }),
    myDocuments: build.query<Document[], void>({
      query: () => ({ url: "documents", params: { scope: "mine" } }),
      providesTags: ["MyDocuments"],
    }),
    documentBySlug: build.query<DocumentWithForms, string>({
      query: slug => ({ url: `documents/by-slug/${encodeURIComponent(slug)}` }),
      providesTags: result => (result ? [{ type: "Document", id: result.id }] : []),
    }),
    documentById: build.query<Document, number>({
      query: id => ({ url: `documents/${id}` }),
      providesTags: (_result, _error, id) => [{ type: "Document", id }],
    }),
    createDocument: build.mutation<CreateDocumentResponse, CreateDocumentRequest>({
      query: body => ({ url: "documents", method: "POST", body }),
      invalidatesTags: ["Documents", "MyDocuments"],
    }),
    updateDocument: build.mutation<SuccessResponse, UpdateDocumentRequest>({
      query: ({ id, ...patch }) => ({ url: `documents/${id}`, method: "PATCH", body: patch }),
      invalidatesTags: (_result, _error, { id }) => [
        "Documents",
        "MyDocuments",
        { type: "Document", id },
      ],
    }),
    deleteDocument: build.mutation<SuccessResponse, number>({
      query: id => ({ url: `documents/${id}`, method: "DELETE" }),
      invalidatesTags: ["Documents", "MyDocuments"],
    }),
    documentAnalytics: build.query<DocumentAnalytics, number>({
      query: id => ({ url: `documents/${id}/analytics` }),
      providesTags: (_result, _error, id) => [{ type: "DocAnalytics", id }],
    }),
    documentSubmissions: build.query<SubmissionWithUser[], number>({
      query: id => ({ url: `documents/${id}/submissions` }),
      providesTags: (_result, _error, id) => [{ type: "DocSubmissions", id }],
    }),
    submitQuiz: build.mutation<SubmitQuizResponse, SubmitQuizRequest>({
      query: body => ({ url: "quiz/submissions", method: "POST", body }),
      invalidatesTags: (_result, _error, { documentId }) => [
        "MySubmissions",
        { type: "DocAnalytics", id: documentId },
        { type: "DocSubmissions", id: documentId },
      ],
    }),
    submitQuizBatch: build.mutation<SubmitQuizBatchResponse, SubmitQuizBatchRequest>({
      query: body => ({ url: "quiz/submissions/batch", method: "POST", body }),
      invalidatesTags: (_result, _error, { documentId }) => [
        "MySubmissions",
        { type: "DocAnalytics", id: documentId },
        { type: "DocSubmissions", id: documentId },
      ],
    }),
    mySubmissions: build.query<SubmissionWithDocument[], void>({
      query: () => ({ url: "quiz/submissions", params: { scope: "mine" } }),
      providesTags: ["MySubmissions"],
    }),
    submissionById: build.query<SubmissionDetail, number>({
      query: id => ({ url: `quiz/submissions/${id}` }),
      providesTags: (_result, _error, id) => [{ type: "Submission", id }],
    }),
  }),
});

export const {
  useListDocumentsQuery,
  useMyDocumentsQuery,
  useDocumentBySlugQuery,
  useDocumentByIdQuery,
  useCreateDocumentMutation,
  useUpdateDocumentMutation,
  useDeleteDocumentMutation,
  useDocumentAnalyticsQuery,
  useDocumentSubmissionsQuery,
  useSubmitQuizMutation,
  useSubmitQuizBatchMutation,
  useMySubmissionsQuery,
  useSubmissionByIdQuery,
} = api;

