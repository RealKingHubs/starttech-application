import { createFileRoute, Link } from '@tanstack/react-router';
import { CheckCircle2, Zap, Shield, ArrowRight } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';

export const Route = createFileRoute('/')({
    component: LandingPage,
});

function LandingPage() {
    return (
        <div className="min-h-screen bg-gradient-to-b from-indigo-50 via-white to-gray-50">
            {/* Hero Section */}
            <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 pt-20 pb-16">
                <div className="text-center">
                    <h1 className="text-5xl md:text-6xl font-bold text-gray-900 mb-6">
                        Organize Your Life with{' '}
                        <span className="text-indigo-600">MuchToDo</span>
                    </h1>
                    <p className="text-xl text-gray-600 mb-8 max-w-2xl mx-auto">
                        A simple, powerful task management app that helps you stay organized and productive.
                        Create, track, and complete your tasks effortlessly.
                    </p>
                    <div className="flex gap-4 justify-center">
                        <Link to="/register">
                            <Button size="lg" className="gap-2">
                                Get Started Free
                                <ArrowRight className="w-5 h-5" />
                            </Button>
                        </Link>
                        <Link to="/login">
                            <Button size="lg" variant="outline">
                                Sign In
                            </Button>
                        </Link>
                    </div>
                </div>
            </div>

            {/* Features Section */}
            <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-16">
                <h2 className="text-3xl font-bold text-center text-gray-900 mb-12">
                    Why Choose MuchToDo?
                </h2>
                <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
                    <Card>
                        <CardContent className="pt-6">
                            <div className="flex flex-col items-center text-center">
                                <div className="p-3 bg-indigo-100 rounded-full mb-4">
                                    <CheckCircle2 className="w-8 h-8 text-indigo-600" />
                                </div>
                                <h3 className="text-xl font-semibold mb-2">Simple & Intuitive</h3>
                                <p className="text-gray-600">
                                    Clean interface that makes task management a breeze. No learning curve required.
                                </p>
                            </div>
                        </CardContent>
                    </Card>

                    <Card>
                        <CardContent className="pt-6">
                            <div className="flex flex-col items-center text-center">
                                <div className="p-3 bg-green-100 rounded-full mb-4">
                                    <Zap className="w-8 h-8 text-green-600" />
                                </div>
                                <h3 className="text-xl font-semibold mb-2">Lightning Fast</h3>
                                <p className="text-gray-600">
                                    Quick load times and snappy interactions. Your tasks are always at your fingertips.
                                </p>
                            </div>
                        </CardContent>
                    </Card>

                    <Card>
                        <CardContent className="pt-6">
                            <div className="flex flex-col items-center text-center">
                                <div className="p-3 bg-purple-100 rounded-full mb-4">
                                    <Shield className="w-8 h-8 text-purple-600" />
                                </div>
                                <h3 className="text-xl font-semibold mb-2">Secure & Private</h3>
                                <p className="text-gray-600">
                                    Your data is encrypted and protected. We take your privacy seriously.
                                </p>
                            </div>
                        </CardContent>
                    </Card>
                </div>
            </div>

            {/* CTA Section */}
            <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-16">
                <Card className="bg-gradient-to-r from-indigo-600 to-purple-600 border-0">
                    <CardContent className="p-12 text-center">
                        <h2 className="text-3xl font-bold text-white mb-4">
                            Ready to Get Organized?
                        </h2>
                        <p className="text-indigo-100 mb-8 text-lg">
                            Join thousands of users who are already managing their tasks more efficiently.
                        </p>
                        <Link to="/register">
                            <Button size="lg" variant="secondary" className="gap-2">
                                Create Your Free Account
                                <ArrowRight className="w-5 h-5" />
                            </Button>
                        </Link>
                    </CardContent>
                </Card>
            </div>

            {/* Footer */}
            <footer className="bg-gray-900 text-gray-400 py-8 mt-16">
                <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 text-center">
                    <p>&copy; 2024 MuchToDo. Built with ❤️ for productivity.</p>
                </div>
            </footer>
        </div>
    );
}