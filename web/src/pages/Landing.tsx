import { Link } from 'react-router-dom';
import { Briefcase, Search, Award } from 'lucide-react';

export function Landing() {
  return (
    <div className="hero min-h-[80vh] bg-base-100">
      <div className="hero-content text-center">
        <div className="max-w-2xl">
          <h1 className="text-5xl font-bold mb-4">Your Career,<br />One Passport</h1>
          <p className="text-lg opacity-70 mb-8">
            Build your complete career profile. Let AI evaluate your skills.
            Get discovered by verified companies. Grow with personalized feedback.
          </p>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-8">
            <div className="card bg-base-200 p-4">
              <Briefcase className="mx-auto mb-2" size={28} />
              <h3 className="font-semibold">Build Profile</h3>
              <p className="text-sm opacity-60">Add every experience — job, gig, project, or education</p>
            </div>
            <div className="card bg-base-200 p-4">
              <Search className="mx-auto mb-2" size={28} />
              <h3 className="font-semibold">Get Discovered</h3>
              <p className="text-sm opacity-60">Verified companies find you by skills and experience</p>
            </div>
            <div className="card bg-base-200 p-4">
              <Award className="mx-auto mb-2" size={28} />
              <h3 className="font-semibold">Grow Faster</h3>
              <p className="text-sm opacity-60">Know your strengths and get suggestions to improve</p>
            </div>
          </div>

          <div className="flex gap-4 justify-center">
            <Link to="/auth/register" className="btn btn-primary btn-lg">Get Started</Link>
            <Link to="/jobs" className="btn btn-outline btn-lg">Browse Jobs</Link>
          </div>
        </div>
      </div>
    </div>
  );
}
