import { Button, Icon, type IconData } from '@gravity-ui/uikit'
import {
  ArrowRotateLeft,
  BookOpen,
  Check,
  Code,
  Comments,
  FileCode,
  Lock,
  MagicWand,
} from '@gravity-ui/icons'
import { useNavigate } from 'react-router-dom'

import productUi from '../assets/product-ui.png'
import { useAuth } from '../hooks/useAuth'
import './landing.css'

export default function LandingPage() {
  const { user } = useAuth()
  const navigate = useNavigate()
  const openAssistant = () => navigate(user ? '/chats' : '/login')

  return (
    <div className="landing-page">
      <header className="landing-topbar">
        <div className="landing-container landing-nav">
          <a href="#top" className="landing-brand" aria-label="BlueSales Bot Assistant">
            <span className="landing-brand-mark">BS</span>
            <span className="landing-brand-copy">
              <span>BlueSales</span>
              <small>Bot Assistant</small>
            </span>
          </a>
          <nav className="landing-nav-links" aria-label="Основная навигация">
            <a href="#features">Возможности</a>
            <a href="#how">Как работает</a>
            <a href="#security">Безопасность</a>
            <a href="#faq">FAQ</a>
          </nav>
          <Button view="action" size="l" onClick={openAssistant}>
            {user ? 'Перейти к чатам' : 'Открыть ассистента'}
          </Button>
        </div>
      </header>

      <main id="top">
        <section className="landing-hero">
          <div className="landing-container landing-hero-grid">
            <div className="landing-hero-copy">
              <div className="landing-eyebrow">
                <span className="landing-eyebrow-dot" /> AI-помощник для BlueSales
              </div>
              <h1>Собирайте ботов через обычный диалог.</h1>
              <p>
                Опишите задачу своими словами. Ассистент опирается на примеры и правила
                BlueSales, помогает собрать логику, проверить условия и подготовить конфигурацию
                бота.
              </p>
              <div className="landing-hero-actions">
                <Button view="action" size="xl" onClick={openAssistant}>
                  Попробовать
                </Button>
                <a className="landing-secondary-action" href="#how">
                  Как это работает
                </a>
              </div>
              <div className="landing-hero-note">
                <Icon data={Lock} size={16} />
                Стабильный контекст: каждый новый чат фиксирует актуальный снимок базы знаний.
              </div>
            </div>

            <div
              className="landing-product-frame"
              aria-label="Скриншот интерфейса BlueSales Bot Assistant"
            >
              <div className="landing-browser-bar">
                <i />
                <i />
                <i />
              </div>
              <img
                src={productUi}
                alt="Интерфейс BlueSales Bot Assistant: чат с AI-помощником и ответом по настройке бота"
              />
              <div className="landing-floating-chip">
                <span className="landing-chip-icon">
                  <Icon data={MagicWand} size={16} />
                </span>
                <span>Ответ с учётом базы знаний</span>
              </div>
            </div>
          </div>
        </section>

        <div className="landing-proof">
          <div className="landing-container landing-proof-inner">
            <div className="landing-proof-copy">
              <strong>Меньше ручной работы с примерами.</strong> Больше времени на сам сценарий
              продаж.
            </div>
            <div className="landing-proof-badges">
              <span>Официальные примеры BlueSales</span>
              <span>Чаты + файлы</span>
              <span>Потоковые ответы</span>
            </div>
          </div>
        </div>

        <section id="features">
          <div className="landing-container">
            <SectionHeading
              kicker="Возможности"
              title="Из бизнес-задачи — в понятную логику бота."
              description="Не нужно помнить синтаксис и искать подходящий пример по документации: ассистент держит базу знаний в контексте разговора."
            />
            <div className="landing-features">
              <FeatureCard icon={MagicWand} title="Создание с нуля">
                Опишите, что должен делать бот: приветствие, распределение, теги, ограничения и
                сценарии повторного запуска — и получите готовую основу.
              </FeatureCard>
              <FeatureCard icon={Code} title="Правки существующих ботов">
                Прикрепите JSON или вставьте фрагмент. Ассистент разберёт текущую логику,
                предложит изменение и объяснит, что именно поменял.
              </FeatureCard>
              <FeatureCard icon={BookOpen} title="Контекст документации">
                Документы базы знаний объединяются в системный контекст, поэтому ответы строятся
                с учётом загруженных правил и примеров.
              </FeatureCard>
              <FeatureCard icon={ArrowRotateLeft} title="Снимки базы знаний">
                Новые чаты получают актуальный снимок базы. Уже начатые диалоги продолжаются на
                том же контексте.
              </FeatureCard>
              <FeatureCard icon={Comments} title="Ответы в реальном времени">
                Текст приходит потоково. Можно продолжать диалог, уточнять детали и шаг за шагом
                доводить сценарий до результата.
              </FeatureCard>
              <FeatureCard icon={FileCode} title="Работа с текстовыми файлами">
                Прикрепляйте JSON, YAML, XML, CSV и другие текстовые форматы, чтобы обсуждать
                реальные настройки.
              </FeatureCard>
            </div>
          </div>
        </section>

        <section className="landing-steps-wrap" id="how">
          <div className="landing-container">
            <SectionHeading
              kicker="Как работает"
              title="Три шага вместо поиска по десяткам примеров."
            />
            <div className="landing-steps">
              <Step number="01" title="Опишите сценарий">
                Например: «сделай так, чтобы один клиент не мог запустить бота два раза».
              </Step>
              <Step number="02" title="Получите решение">
                Ассистент подберёт условия и действия из базы знаний и предложит конкретное
                изменение.
              </Step>
              <Step number="03" title="Уточните и примените">
                Задайте дополнительные условия, приложите текущий файл и доведите конфигурацию до
                рабочего варианта.
              </Step>
            </div>
          </div>
        </section>

        <section>
          <div className="landing-container landing-use-grid">
            <div className="landing-use-panel">
              <div className="landing-kicker">Примеры запросов</div>
              <h2>Можно писать как человеку.</h2>
              <div className="landing-prompt-list">
                <div>«Сделай распределение новых клиентов между менеджерами по кругу.»</div>
                <div>«Не запускай приветствие повторно, если клиент уже проходил сценарий.»</div>
                <div>«Вот мой JSON. Добавь тег после успешного шага и объясни изменения.»</div>
                <div>«Проверь, почему это правило может срабатывать дважды.»</div>
              </div>
            </div>
            <aside className="landing-dark-panel">
              <h3>Не просто чат. Ассистент с вашей базой знаний.</h3>
              <p>
                Перед отправкой запроса модель получает собранный контекст с документами и
                инструкциями.
              </p>
              <div className="landing-mini-list">
                {[
                  'Стабильный порядок документов в контексте',
                  'Контроль актуальности базы знаний',
                  'Отдельный контекст для каждого нового чата',
                ].map((item) => (
                  <div className="landing-mini-item" key={item}>
                    <span>
                      <Icon data={Check} size={14} />
                    </span>
                    {item}
                  </div>
                ))}
              </div>
            </aside>
          </div>
        </section>

        <section id="security">
          <div className="landing-container landing-security">
            <div className="landing-security-copy">
              <div className="landing-kicker">Техническая основа</div>
              <h2>Сделано как рабочий сервис, а не демо.</h2>
              <p>
                Авторизация, база данных, снимки знаний, потоковые ответы и серверная часть уже
                собраны в полноценный SPA-сервис.
              </p>
            </div>
            <div className="landing-security-grid">
              <SecurityItem title="Пароли">Хранятся как bcrypt-хэши.</SecurityItem>
              <SecurityItem title="Сессии">Токен хранится в httpOnly cookie.</SecurityItem>
              <SecurityItem title="База знаний">Снимки фиксируются по SHA-256.</SecurityItem>
              <SecurityItem title="Поток ответов">SSE разделяет текст и статистику.</SecurityItem>
            </div>
          </div>
        </section>

        <section id="faq">
          <div className="landing-container">
            <SectionHeading kicker="FAQ" title="Частые вопросы" />
            <div className="landing-faq">
              <Faq title="Что именно умеет ассистент?" open>
                Помогает создавать и изменять сценарии ботов, разбирать конфигурации, подбирать
                условия и объяснять внесённые изменения.
              </Faq>
              <Faq title="На чём основаны ответы?">
                На системных инструкциях и базе знаний с официальными примерами и внутренними
                материалами.
              </Faq>
              <Faq title="Можно ли прикрепить готовый файл бота?">
                Да. Сервис принимает текстовые вложения, включая JSON, YAML, XML, CSV и TXT.
              </Faq>
              <Faq title="Что происходит после обновления базы знаний?">
                Новые чаты получают новый снимок базы, а начатые продолжают работать с ранее
                закреплённым снимком.
              </Faq>
            </div>
          </div>
        </section>

        <section className="landing-cta" id="start">
          <div className="landing-container landing-cta-box">
            <div>
              <h2>Есть задача для бота? Опишите её.</h2>
              <p>
                Откройте ассистента, создайте новый чат и напишите, что должно происходить с
                клиентом. Начать можно с одной фразы.
              </p>
            </div>
            <Button view="action" size="xl" onClick={openAssistant}>
              Открыть BlueSales Bot Assistant
            </Button>
          </div>
        </section>
      </main>

      <footer className="landing-footer">
        <div className="landing-container landing-footer-inner">
          <div className="landing-brand">
            <span className="landing-brand-mark">BS</span>
            <span className="landing-brand-copy">
              <span>BlueSales</span>
              <small>Bot Assistant</small>
            </span>
          </div>
          <div>AI-помощник для настройки и создания ботов в BlueSales</div>
        </div>
      </footer>
    </div>
  )
}

function SectionHeading({
  kicker,
  title,
  description,
}: {
  kicker: string
  title: string
  description?: string
}) {
  return (
    <div className="landing-section-head">
      <div className="landing-kicker">{kicker}</div>
      <h2>{title}</h2>
      {description && <p>{description}</p>}
    </div>
  )
}

function FeatureCard({
  icon,
  title,
  children,
}: {
  icon: IconData
  title: string
  children: string
}) {
  return (
    <article className="landing-card">
      <div className="landing-icon">
        <Icon data={icon} size={22} />
      </div>
      <h3>{title}</h3>
      <p>{children}</p>
    </article>
  )
}

function Step({ number, title, children }: { number: string; title: string; children: string }) {
  return (
    <article className="landing-step">
      <div className="landing-step-num">ШАГ {number}</div>
      <h3>{title}</h3>
      <p>{children}</p>
    </article>
  )
}

function SecurityItem({ title, children }: { title: string; children: string }) {
  return (
    <div className="landing-security-item">
      <strong>{title}</strong>
      <span>{children}</span>
    </div>
  )
}

function Faq({ title, children, open = false }: { title: string; children: string; open?: boolean }) {
  return (
    <details open={open}>
      <summary>{title}</summary>
      <p>{children}</p>
    </details>
  )
}
