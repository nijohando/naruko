(defproject jp.nijohando.naruko/naruko-slack-notifier "1.0.0"
  :description "AWS Lambda for loading data into AWS Firehose"
  :license {:name "Eclipse Public License"
            :url "http://www.eclipse.org/legal/epl-v10.html"}
  :dependencies [[org.clojure/clojure "1.9.0"]
                 [org.clojure/clojurescript "1.9.946"]
                 [org.clojure/spec.alpha "0.1.143"]
                 [org.clojure/test.check "0.10.0-alpha2"]
                 [org.clojure/core.async "0.4.474"]
                 [jp.nijohando/failable "0.1.1"]
                 [jp.nijohando/deferable "0.1.0"]
                 [funcool/httpurr "1.0.0"]
                 [funcool/promesa "1.9.0"]]
  :plugins [[lein-figwheel "0.5.14"]]
  :source-paths ["src/main/clj", "src/dev/clj", "src/main/cljs"]
  :clean-targets [:target-path "out"]
  :aliases {"fwrepl" ["run" "-m" "clojure.main" "src/dev/clj/tools/repl.clj"]
            "build" ["run" "-m" "clojure.main" "src/dev/clj/tools/build.clj"]
            "package" ["run" "-m" "clojure.main" "src/dev/clj/tools/package.clj"]}
  :profiles {:dev {:dependencies [[figwheel-sidecar "0.5.14"]
                                  [com.cemerick/piggieback "0.2.1"]]
                   :repl-options {:nrepl-middleware [cemerick.piggieback/wrap-cljs-repl]}}})
