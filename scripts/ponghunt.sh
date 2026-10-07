#!/usr/bin/env bash
# Hunt until a deadline (12:45 PM ET today). Prowls diverse themes on the open web
# and casts the native-tongue news spells, retraining the cerebellum periodically.
# Every engram auto-carries its "akin to" associations, so the connected vector
# space keeps growing. Self-terminates at the target.
set -uo pipefail
cd "$(dirname "$0")/.."
export PATH="$HOME/.local/go/bin:$PATH"

TARGET=$(TZ="America/New_York" date -d 'today 12:45' +%s)
go build -o bin/cerebrum ./cmd/cerebrum

# Pure Arabic word bank (3x density, advanced academic topics)
themes=(
    عقل تفكير إدراك ذاكرة إدراك لغة تعلم تطور وراثة علوم_عصبية وعي حوسبة خوارزمية معلومات إنتروبيا
    ديناميكا_حرارية نسبية كمي كيمياء بروتين بيئة مناخ محيط جيولوجيا فلك مجرة رياضيات طوبولوجيا
    تشفير شبكة عمارة هندسة روبوتات مواد طاقة فلسفة منطق لسانيات موسيقى تاريخ
    فيزياء أحياء إحصاء حساب_تفاضلي جبر_خطي معادلات_تفاضلية تعلم_آلي تعلم_عميق
    ذكاء_اصطناعي علم_البيانات معلوماتية_حيوية تكنولوجيا_نانوية فيزياء_فلكية علم_الكون فيزياء_الجسيمات
    كيمياء_عضوية كيمياء_حيوية أحياء_جزيئية مناعة أحياء_مجهرية علوم_بيئية استدامة طاقة_متجددة حوسبة_كمومية
    أمن_سيبراني سلسلة_الكتل نظم_موزعة حوسبة_سحابية بيانات_كبيرة إنترنت_الأشياء علم_نفس علم_اجتماع
    أنثروبولوجيا علم_سياسة اقتصاد اقتصاد_كلي اقتصاد_جزئي علاقات_دولية أدب_مقارن دراسات_ثقافية
    دراسات_جنسانية ما_بعد_استعمار عولمة تخطيط_حضري سياسة_عامة أخلاقيات معرفة ما_وراء_الطبيعة
    جمالية سيميائية تلحين تأليف_موسيقي نظرية_موسيقية تاريخ_فن ثقافة_بصرية دراسات_سينمائية إنسانيات_رقمية
    علم_إدراك لسانيات_حاسوبية اقتصاد_عصبي أخلاقيات_حيوية اتصال_علمي تصور_بيانات تفاعل_إنسان_حاسوب
    أحياء_نظم أحياء_اصطناعية
)

spells=(العربية الفارسية العبرية الأردية التركية)

start=$(date +%s)
i=0
while [ "$(date +%s)" -lt "$TARGET" ]; do
  theme="${themes[$((i % ${#themes[@]}))]}"
  echo "── [$(TZ=America/New_York date +%H:%M) ET · hunt $i] prowl \"$theme\" ──"
  timeout 220 ./bin/cerebrum prowl "$theme" 2>&1 | sed 's/\x1b\[[0-9;]*m//g' | grep -E "hop|prowl done" || true

  if (( i % 4 == 3 )); then
    sp="${spells[$(((i/4) % ${#spells[@]}))]}"
    echo "── cast $sp (native news) ──"
    timeout 120 ./bin/cerebrum cast "$sp" >/dev/null 2>&1 || true
  fi
  if (( i % 8 == 7 )); then
    ./bin/cerebrum train >/dev/null 2>&1 || true
    echo "   ↻ retrained · bank $(./bin/cerebrum coverage 2>/dev/null | sed 's/\x1b\[[0-9;]*m//g')"
  fi
  i=$((i+1))
done

echo "🐾 deadline reached (12:45 PM ET). final retrain…"
./bin/cerebrum train 2>&1 | sed 's/\x1b\[[0-9;]*m//g' | tail -2
echo "🐾 long hunt done after $i hunts, $(( ($(date +%s)-start)/60 )) min."
./bin/cerebrum coverage 2>&1 | sed 's/\x1b\[[0-9;]*m//g'
