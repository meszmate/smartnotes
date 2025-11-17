import * as React from 'react';
import { Checkbox } from '@/components/ui/checkbox';
import { Label } from '@/components/ui/label';
import * as pdfjs from 'pdfjs-dist';
import mammoth from 'mammoth';
import {
    InputGroup,
    InputGroupAddon,
    InputGroupButton,
    InputGroupTextarea,
} from './ui/input-group';
import { FileUpIcon, X } from 'lucide-react';
import { Spinner } from './ui/spinner';
import { formatBytes } from '@/lib/formatBytes';
import { Turnstile } from '@marsidev/react-turnstile';
import {
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';

export interface AIInputData {
    text: string;
    options: { summary: boolean; flashcards: boolean; quiz: boolean };
    turnstile: string;
}

interface AIInputProps {
    onSubmit: (data: AIInputData) => Promise<void>;
}

export function AIInput({ onSubmit }: AIInputProps) {
    const [inputText, setInputText] = React.useState('');
    const [files, setFiles] = React.useState<FileList | null>(null);
    const [options, setOptions] = React.useState({
        summary: true,
        flashcards: true,
        quiz: true,
    });
    const [isLoading, setIsLoading] = React.useState(false);
    const [showTurnstile, setShowTurnstile] = React.useState(false);
    const [turnstileToken, setTurnstileToken] = React.useState<string | null>(null);

    const fileInputRef = React.useRef<HTMLInputElement>(null);
    const turnstileRef = React.useRef<any>(null);

    const handleOptionChange = (opt: keyof typeof options) => {
        setOptions((p) => ({ ...p, [opt]: !p[opt] }));
    };

    const removeFile = (index: number) => {
        if (!files) return;
        const dt = new DataTransfer();
        Array.from(files)
            .filter((_, i) => i !== index)
            .forEach((f) => dt.items.add(f));
        setFiles(dt.files);
    };

    const openFilePicker = () => fileInputRef.current?.click();

    const extractTextFromFile = async (file: File): Promise<string> => {
        const fileType = file.type;
        const fileName = file.name.toLowerCase();
        if (fileType === 'text/plain' || fileName.endsWith('.txt')) {
            return new Promise((res, rej) => {
                const r = new FileReader();
                r.onload = (e) => res(e.target?.result as string);
                r.onerror = rej;
                r.readAsText(file);
            });
        }
        if (fileType === 'application/pdf' || fileName.endsWith('.pdf')) {
            return new Promise((res, rej) => {
                const r = new FileReader();
                r.onload = async (e) => {
                    try {
                        const buf = e.target?.result as ArrayBuffer;
                        const pdf = await pdfjs.getDocument({ data: new Uint8Array(buf) }).promise;
                        let txt = '';
                        for (let i = 1; i <= pdf.numPages; i++) {
                            const page = await pdf.getPage(i);
                            const content = await page.getTextContent();
                            txt += content.items.map((it: any) => ('str' in it ? it.str : '')).join(' ') + '\n';
                        }
                        res(txt);
                    } catch (err) {
                        rej(err);
                    }
                };
                r.onerror = rej;
                r.readAsArrayBuffer(file);
            });
        }
        if (
            fileType ===
            'application/vnd.openxmlformats-officedocument.wordprocessingml.document' ||
            fileName.endsWith('.docx')
        ) {
            return new Promise((res, rej) => {
                const r = new FileReader();
                r.onload = async (e) => {
                    try {
                        const buf = e.target?.result as ArrayBuffer;
                        const { value } = await mammoth.extractRawText({ arrayBuffer: buf });
                        res(value);
                    } catch (err) {
                        rej(err);
                    }
                };
                r.onerror = rej;
                r.readAsArrayBuffer(file);
            });
        }
        throw new Error(`Unsupported file type: ${file.type}`);
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!inputText.trim() && (!files || files.length === 0)) return;

        // Show Turnstile if not verified
        if (!turnstileToken) {
            setShowTurnstile(true);
            return;
        }

        setIsLoading(true);
        try {
            let combined = inputText.trim();
            if (files?.length) {
                const allowed = Array.from(files).filter((f) => {
                    const ext = f.name.split('.').pop()?.toLowerCase();
                    return ['txt', 'pdf', 'docx'].includes(ext ?? '');
                });
                const texts = await Promise.all(
                    allowed.map((f) =>
                        extractTextFromFile(f).catch((err) => {
                            console.warn('Skip file', f.name, err);
                            return '';
                        })
                    )
                );
                const nonEmpty = texts.filter((t) => t.trim());
                if (nonEmpty.length) {
                    combined += (combined ? '\n\n---\n\n' : '') + nonEmpty.join('\n\n---\n\n');
                }
            }
            if (!combined.trim()) return;

            await onSubmit({
                text: combined,
                options,
                turnstile: turnstileToken,
            });
        } catch (err) {
            console.error(err);
        } finally {
            setIsLoading(false);
            setTurnstileToken(null);
            turnstileRef.current?.reset?.();
        }
    };

    return (
        <>
            <form onSubmit={handleSubmit} className="w-full mx-auto p-4 sm:p-6">
                <InputGroup className="w-full rounded-3xl sm:rounded-4xl p-5 bg-white shadow-sm">
                    <InputGroupTextarea
                        id="text-input"
                        placeholder="Paste your text here, or upload files below..."
                        value={inputText}
                        onChange={(e) => setInputText(e.target.value)}
                        className="min-h-32 resize-none"
                    />

                    <InputGroupAddon
                        align="block-end"
                        className="flex flex-col gap-4 p-4 sm:p-0 sm:flex-row sm:items-center sm:gap-6"
                    >
                        <input
                            ref={fileInputRef}
                            type="file"
                            multiple
                            accept=".txt,.pdf,.docx"
                            onChange={(e) => setFiles(e.target.files)}
                            className="hidden"
                        />

                        <InputGroupButton
                            variant="outline"
                            onClick={openFilePicker}
                            className="rounded-full h-12 w-12 flex items-center justify-center shrink-0"
                            asChild
                        >
                            <span>
                                <FileUpIcon className="h-5 w-5" />
                                <span className="sr-only">Upload files</span>
                            </span>
                        </InputGroupButton>

                        {files && files.length > 0 && (
                            <div className="flex-1 min-w-0">
                                <p className="text-xs text-muted-foreground text-center sm:text-left mb-1">
                                    {files.length} file{files.length > 1 ? 's' : ''} selected
                                </p>
                                <div className="flex flex-wrap justify-center sm:justify-start gap-1.5 max-h-24 overflow-y-auto">
                                    {Array.from(files).map((f, i) => (
                                        <div
                                            key={i}
                                            className="flex items-center gap-1 bg-muted/50 rounded-md px-2 py-1 text-xs whitespace-nowrap"
                                        >
                                            <span className="truncate max-w-28 sm:max-w-40">{f.name}</span>
                                            <span className="text-muted-foreground">({formatBytes(f.size)})</span>
                                            <button
                                                type="button"
                                                onClick={() => removeFile(i)}
                                                className="ml-1 text-destructive hover:text-destructive/80"
                                                aria-label={`Remove ${f.name}`}
                                            >
                                                <X className="h-3 w-3" />
                                            </button>
                                        </div>
                                    ))}
                                </div>
                            </div>
                        )}

                        <div className="flex flex-col gap-3 sm:gap-2">
                            <Label className="text-sm w-full font-medium text-center">Generate</Label>
                            <div className="flex flex-wrap gap-4 sm:gap-6">
                                {(['summary', 'flashcards', 'quiz'] as const).map((k) => (
                                    <div key={k} className="flex items-center space-x-2">
                                        <Checkbox
                                            id={k}
                                            checked={options[k]}
                                            onCheckedChange={() => handleOptionChange(k)}
                                        />
                                        <Label htmlFor={k} className="cursor-pointer capitalize text-sm">
                                            {k === 'quiz' ? 'Quiz Questions' : k}
                                        </Label>
                                    </div>
                                ))}
                            </div>
                        </div>

                        <InputGroupButton
                            variant={"default"}
                            type="submit"
                            disabled={isLoading || (!inputText.trim() && (!files || files.length === 0))}
                            className="w-full sm:w-auto ml-auto rounded-full h-12 px-6 text-base flex items-center justify-center gap-2"
                        >
                            {isLoading && <Spinner className="h-4 w-4" />}
                            Generate
                            <span className="sr-only">Send</span>
                        </InputGroupButton>
                    </InputGroupAddon>
                </InputGroup>
            </form>

            {/* Turnstile Modal */}
            <Dialog open={showTurnstile} onOpenChange={setShowTurnstile}>
                <DialogContent className="sm:max-w-md">
                    <DialogHeader>
                        <DialogTitle>Verify you're human</DialogTitle>
                    </DialogHeader>
                    <div className="flex justify-center py-6">
                        <Turnstile
                            ref={turnstileRef}
                            siteKey={import.meta.env.VITE_TURNSTILE_SITE_KEY}
                            onSuccess={(token) => {
                                setTurnstileToken(token);
                                setShowTurnstile(false);
                                // Auto-submit after verification
                                setTimeout(() => {
                                    const form = document.querySelector('form');
                                    form?.requestSubmit();
                                }, 300);
                            }}
                            onError={() => {
                                setTurnstileToken(null);
                            }}
                            onExpire={() => {
                                setTurnstileToken(null);
                            }}
                        />
                    </div>
                    <div className="flex justify-end">
                        <Button
                            variant="ghost"
                            onClick={() => {
                                setShowTurnstile(false);
                                setTurnstileToken(null);
                                turnstileRef.current?.reset?.();
                            }}
                        >
                            Cancel
                        </Button>
                    </div>
                </DialogContent>
            </Dialog>
        </>
    );
}
