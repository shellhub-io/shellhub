import { Routes, Route } from "react-router-dom";
import { ScrollToTop } from "@/components";
import { routes } from "@/routes";

/**
 * Root of the marketing site. Every route is declared in routes, so a page is added there and
 * nowhere else; ScrollToTop sits outside the switch so a navigation always lands at the top.
 */
export default function App() {
  return (
    <>
      <ScrollToTop />
      <Routes>
        {routes.map((r) => (
          <Route key={r.path} path={r.path} element={r.element} />
        ))}
      </Routes>
    </>
  );
}
