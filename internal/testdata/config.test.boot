blacklist {
        disabled false
        dns-redirect-ip 192.168.168.1
        domains {
            source tasty {
                description "File source"
                dns-redirect-ip 10.10.10.10
                file ../../internal/testdata/blist.hosts.src
            }
        }
        exclude 1e100.net
        exclude 2o7.net
        exclude adobedtm.com
        exclude akamai.net
        exclude akamaihd.net
        exclude amazon.com
        exclude amazonaws.com
        exclude apple.com
        exclude ask.com
        exclude avast.com
        exclude avira-update.com
        exclude bannerbank.com
        exclude bing.com
        exclude bit.ly
        exclude bitdefender.com
        exclude cdn.ravenjs.com
        exclude cdn.visiblemeasures.com
        exclude cloudfront.net
        exclude coremetrics.com
        exclude dropbox.com
        exclude ebay.com
        exclude edgesuite.net
        exclude evernote.com
        exclude express.co.uk
        exclude feedly.com
        exclude freedns.afraid.org
        exclude github.com
        exclude githubusercontent.com
        exclude global.ssl.fastly.net
        exclude google.com
        exclude googleads.g.doubleclick.net
        exclude googleadservices.com
        exclude googleapis.com
        exclude googletagmanager.com
        exclude googleusercontent.com
        exclude gstatic.com
        exclude gvt1.com
        exclude gvt1.net
        exclude hb.disney.go.com
        exclude herokuapp.com
        exclude hp.com
        exclude hulu.com
        exclude images-amazon.com
        exclude live.com
        exclude magnetmail1.net
        exclude microsoft.com
        exclude microsoftonline.com
        exclude msdn.com
        exclude msecnd.net
        exclude msftncsi.com
        exclude mywot.com
        exclude nsatc.net
        exclude paypal.com
        exclude pop.h-cdn.co
        exclude rackcdn.com
        exclude rarlab.com
        exclude schema.org
        exclude shopify.com
        exclude skype.com
        exclude smacargo.com
        exclude sourceforge.net
        exclude spotify.com
        exclude spotify.edgekey.net
        exclude spotilocal.com
        exclude ssl-on9.com
        exclude ssl-on9.net
        exclude sstatic.net
        exclude static.chartbeat.com
        exclude storage.googleapis.com
        exclude twimg.com
        exclude viewpoint.com
        exclude windows.net
        exclude xboxlive.com
        exclude yimg.com
        exclude ytimg.com
        hosts {
            include beap.gemini.yahoo.com
            source hageziPro {
                description "HaGeZi DNS Blocklists — Pro (dnsmasq)"
                prefix ""
                url https://cdn.jsdelivr.net/gh/hagezi/dns-blocklists@latest/dnsmasq/pro.txt
            }
        }
    }
