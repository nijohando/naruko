(ns tools
  (:require [cljs.repl]
            [cljs.repl.node]
            [cljs.build.api]
            [clojure.java.io :as jio]
            [jp.nijohando.deferable :refer [do* defer]]
            [figwheel-sidecar.repl-api :as ra])
  (:import (java.io File)
           (java.nio.file Path
                          FileVisitOption
                          Files)
           (java.util Comparator)
           (java.util.zip ZipEntry
                          ZipOutputStream)))

(def archive-file-name "naruko-slack-notifier.zip")
(def cljs-src-dir "src/main/cljs")
(def target-dir "target")
(def cljs-out-dir (str target-dir "/out"))
(def cljs-main-js "index.js")
(def cljs-main-ns "jp.nijohando.naruko.slack-notifier.core")
(def cljs-output-to (str cljs-out-dir "/" cljs-main-js))
(def cljs-source-map (str cljs-output-to ".map"))
(def cljs-npm-deps {})
(def cljs-dev-npm-deps {:ws "4.0.0"})
(def cljs-compiler-opts {:output-dir cljs-out-dir
                         :output-to cljs-output-to
                         :optimizations :simple
                         :source-map cljs-source-map
                         :npm-deps cljs-npm-deps
                         :install-deps true
                         :target :nodejs
                         :verbose true})
(def cljs-dev-compiler-opts (merge cljs-compiler-opts {
                                   :main cljs-main-ns
                                   :npm-deps cljs-dev-npm-deps
                                   :optimizations :none
                                   :source-map true}))

(defn- rmdir [fname]
  (letfn [(del [file]
            (when (.isDirectory file)
              (doseq [child-file (.listFiles file)]
                (del child-file)))
            (clojure.java.io/delete-file file))]
    (let [f (clojure.java.io/file fname)]
      (when (.exists f)
        (del f)))))

(defn- clean-node-modules []
  (time (rmdir "node_modules")))

(defn build
  []
  (->> cljs-compiler-opts
       (cljs.build.api/build cljs-src-dir)))

(defn package
  []
  (clean-node-modules)
  (build)
  (do*
    (let [zip (ZipOutputStream. (jio/output-stream (str target-dir "/" archive-file-name)))
           _ (defer (.close zip))
           main-js (jio/file cljs-output-to)
           source-map (jio/file cljs-source-map)
           append (fn [^String entry-name ^File f]
                    (.putNextEntry zip (ZipEntry. entry-name))
                    (jio/copy f zip)
                    (.closeEntry zip))]
      (append (.getName main-js) main-js)
      (append (.getName source-map) source-map)
      (doseq [f (file-seq (jio/file "node_modules")) :when (.isFile f)]
        (append (.getPath f) f)))))

(defn repl
  []
  (ra/start-figwheel!
    {:figwheel-options {} ;; <-- figwheel server config goes here 
     :build-ids ["dev"]   ;; <-- a vector of build ids to start autobuilding
     :all-builds          ;; <-- supply your build configs here
     [{:id "dev"
       :figwheel true
       :source-paths [cljs-src-dir]
       :compiler cljs-dev-compiler-opts}]})
  (ra/cljs-repl))

