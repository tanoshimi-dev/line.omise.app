import Header from '@/components/Header'
import HeroSection from '@/components/HeroSection'
import DemoAppsSection from '@/components/DemoAppsSection'
import FeaturesSection from '@/components/FeaturesSection'
import WhyUsSection from '@/components/WhyUsSection'
import ContactSection from '@/components/ContactSection'
import ProfileSection from '@/components/ProfileSection'
import Footer from '@/components/Footer'

export default function App() {
  return (
    <div className="min-h-screen">
      <Header />
      <main>
        <HeroSection />
        <DemoAppsSection />
        <FeaturesSection />
        <WhyUsSection />
        <ContactSection />
        <ProfileSection />
      </main>
      <Footer />
    </div>
  )
}
