export default function ProfileSection() {
  return (
    <section id="profile" className="py-20 sm:py-28 bg-gradient-to-b from-gray-50 to-white">
      <div className="mx-auto max-w-3xl px-4 sm:px-6 lg:px-8 text-center">
        {/* Profile Image */}
        <div className="mx-auto mb-3 h-32 w-32 overflow-hidden rounded-full shadow-lg ring-4 ring-white">
          <img
            src="/images/profile.png"
            alt="プロフィール"
            className="h-full w-full object-cover"
          />
        </div>

        {/* Introduction */}
        <p className="text-gray-600">
          はんなりdevの公式アカウントです。
        </p>
        <p className="mt-2 inline-flex items-center gap-1.5 text-gray-600">
          まずは友だち追加お願いします！
          <span className="text-xl animate-[wiggle_1s_ease-in-out_infinite]">👋</span>
        </p>
        {/* LINE QR Code */}
        <div className="my-4">
          <img
            src="https://qr-official.line.me/gs/M_497cqdxt_BW.png?oat_content=qr"
            alt="LINE QR Code"
            className="mx-auto w-48 h-48"
          />
        </div>
        {/* Career & Profile */}
        <div className="mt-12 rounded-2xl bg-white shadow-lg ring-1 ring-gray-100 p-8 sm:p-10 text-left">
          <h3 className="text-lg font-bold text-gray-900 flex items-center gap-2">
            <span className="inline-flex items-center justify-center h-8 w-8 rounded-lg bg-line-green/10">
              <svg className="h-5 w-5 text-line-green" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
              </svg>
            </span>
            プロフィール
          </h3>

          <p className="mt-4 text-gray-600 leading-relaxed">
            <div>
            大手SIer・社内SEを経て、フリーランスエンジニアとして活動しています。
            </div>
            <div>
            キャリア20年のエンジニアが、あなたの店舗のDXを直接サポートします。
            </div>
          </p>

          {/* Career highlights */}
          <div className="mt-6 grid grid-cols-1 sm:grid-cols-3 gap-4">
            <div className="rounded-xl bg-gray-50 p-4 text-center">
              <div className="text-2xl font-bold text-line-green">15年</div>
              <div className="mt-1 text-sm text-gray-500">SIer経験</div>
            </div>
            <div className="rounded-xl bg-gray-50 p-4 text-center">
              <div className="text-2xl font-bold text-line-green">5年</div>
              <div className="mt-1 text-sm text-gray-500">社内SE経験</div>
            </div>
            <div className="rounded-xl bg-gray-50 p-4 text-center">
              <div className="text-2xl font-bold text-line-green">20年+</div>
              <div className="mt-1 text-sm text-gray-500">トータルキャリア</div>
            </div>
          </div>


        </div>
      </div>
    </section>
  )
}
