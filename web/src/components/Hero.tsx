import React from 'react';
import { Badge } from '@/components/ui/badge';
import { AIInput } from './Prompt';
import { Sparkles } from 'lucide-react';
import Animate from './Animate';
import axios, { isAxiosError } from "axios"
import toast from 'react-hot-toast';
import type { Data } from '@/lib/types';
import { useData } from '@/lib/dataContext';
import type { AIInputData } from "./Prompt"
import ProcessingModal from './ProcessingModal';
import { useNavigate } from 'react-router-dom';

interface ApiErrorResponse {
    error: string;
    message: string;
}

const HeroSection: React.FC = () => {
    const navigate = useNavigate();
    const datas = useData()
    const [isLoading, setIsLoading] = React.useState<boolean>(false);

    async function onSubmit(data: AIInputData) {
        setIsLoading(true)
        try {
            const resp = await axios.post<Data>(
                `${import.meta.env.VITE_API_URL}/generate`,
                {
                    prompt: data.text,
                    ...data.options,
                    turnstile: data.turnstile,
                },
                {
                    headers: { 'Content-Type': 'application/json' },
                }
            );

            // Success (2xx)
            const newEntry: Data = resp.data;
            datas.updateData([...datas.data, newEntry]);

            toast.success('Your study material is ready!');
            navigate(`/${newEntry.id}`)
        } catch (err) {
            // -------------------------------------------------
            //  Axios network / 4xx / 5xx errors
            // -------------------------------------------------
            if (isAxiosError<ApiErrorResponse>(err)) {
                const status = err.status ?? err.response?.status;

                // API returned { error, message }
                const apiError = err.response?.data;
                const errorMsg = apiError?.message ?? err.message;

                // 400 – client error
                if (status === 400) {
                    toast.error(`Invalid request: ${errorMsg}`)
                    return;
                }

                // 500 – server error
                if (status && status >= 500) {
                    toast.error(`Server error: ${errorMsg}`)
                    return;
                }

                // Any other 4xx (401, 403, etc.)
                toast.error(`Error (${status}): ${errorMsg}`)
            } else {
                // Network error, timeout, etc.
                toast.error(`Network Error\nPlease check your connection and try again.`)
            }
        } finally {
            setIsLoading(false)
        }
    }
    return (
        <section className="relative min-h-screen overflow-hidden">
            <div className="relative max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 pt-20 pb-16">
                <div className="text-center mb-12">
                    <Badge variant="secondary" className="mb-4 bg-neutral-50 border border-neutral-300 text-neutral-800">
                        <Sparkles className="w-4 h-4 mr-1" />
                        AI-Powered Learning
                    </Badge>

                    <h1 className="text-6xl font-semibold text-neutral-900 mb-6">
                        Transform Your Study Materials
                    </h1>

                    <p className="text-xl text-gray-600 max-w-3xl mx-auto">
                        Upload your notes, textbooks, or any study material. Our AI will create
                        summaries, flashcards, and quiz questions to supercharge your learning.
                    </p>
                </div>

                <Animate>
                    <AIInput onSubmit={onSubmit} />

                </Animate>

                <ProcessingModal isOpen={isLoading} />
            </div>
        </section>
    );
};

export default HeroSection;
