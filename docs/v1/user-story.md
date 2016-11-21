# A Year In The Life Of A Kubernetes Service Developer

## 0. Today
I'm a developer in an enterprise environment.  For some fun in my spare time, I’m building a node.js chat application for me 
and a couple of coworkers to goof around with.  It’s really simple.  It runs on one container.  To get going, I simply 
download a node.js container onto minikube on my laptop and started coding.  I want some co-workers to be able to find the 
chat app to connect.  So, I register it as a service in service in the kubernetes service catalog.  It keeps track of the IP 
address and the port of the app for us which is nice because I keep moving the thing around.  It started on minikube on my 
laptop, but now it’s in the company’s kubernetes deployment on our OpenStack private cloud.

## 1. Bind kubernetes whitebox services to an internal blackbox service
The folks I had invited to use the app and were having fun with it and told people from another department about it and now 
some of them want to join.  With all these people using the app, it’s more of a problem when a message is dropped.  Every 
time the system glitches or a client loses connection, the message is lost to the ether.  I really should connect this to a 
message queue so that I can make sure that messages are not deleted on the server side until I’m assured they’ve been 
delivered.  I hoping to avoid the work of picking a system and learning it.  The good news is that when I was looking at the 
kubernetes service catalog the other day, I saw a RabbitMQ instance that someone else was maintaining.  All I needed to do 
was issue a single command to bind my chat app to the Rabbit service (and, of course, edit my software to use Rabbit).  The 
bind call gave back credentials which were automatically discovered by my chat app and it was off to the races.  I have no 
idea where that thing is running or how the firewalls work, but whatever, I trust it since it’s maintained by corporate IT.

## 2. Make a type from a kubernetes service
## 3. Share service types inside an org 
## 4. Deploy service instance from a service types inside an org 	
To my surprise, this app started to really take off.  People were having so much fun with it that it became part of company 
morale.  I was assigned maintaining the app as a 20% project and even have a couple interns working with me on bug fixing.  
I now get paid to play with my service which is cool, but now I’m starting to get support tickets for it.

Now I need to maintain multiple instances of this.  I have test and staging instances.  I also have two to three versions 
under development at any given time.  I can deploy this thing with my eyes closed, but the interns have been accidentally 
creating some difficulties for me.  They sometimes do strange configuration of test instances.  When I ask them to be be 
more mindful about deployment, they push back on me.  They tell me that the company has adopted a standardized way to 
templatize kubernetes hosted software that (a) handles parameterization so they know what they can play with and what I want 
them to leave alone and (b) creates a standard provisioning API so that they can deploy my chat app the same way that they 
deploy everything else in the company.  I check it out and it’s pretty cool - creating the template and parameters is 
<simple process> and I can deploy the app with a single command.  The kubernetes UI even does the work of taking the 
argument list and making fields for me in the deployment screen.  Even though it was not hard to deploy for me before, it’s 
easier now even (whether from CLI or UI) and the interns are always deploying within the constraints I set.

## 5. Attach a standard binding template
## 6. Share service instances inside an org
Someone in tech support just had the idea that they can use my chat app inside the call center system.  They need to connect 
their app to my chat service the same way that I connected to the Rabbit service that IT maintains for me.  I call up the 
Rabbit person and ask how they made that binding thing work.  Luckily, it was not hard at all.  I just did <simple workflow> 
to make a service broker to respond to binding requests, deployed an instance and the customer support people were off to 
the races.

## 7. Buy access to a type from the marketplace  
## 8. Bind a kubernetes service to a service outside kubernetes  	
Those tech support people are expecting something with a bit more enterprise capability.  They want the app to persist 
information about chat logs and are are very concerned about the times we’ve lost it and brought it back empty.  This app 
was just supposed to be for fun; I was not thinking about customer satisfaction audits.  If I could bring in a well managed 
database without hiring a DBA, life would be great.  A friend of mine works at a database company that sells and manages 
databases on the public cloud that we use.  She told me that if I was using Kubernetes or Cloud Foundry, this was trivial to 
solve.  They implemented the exact same service broker model I had to implement.  I just needed to buy their database from 
some store and with one API call they will deploy a managed database instance to the same cloud as my chat app and with a 
second API call I can bind my app to it.  I need to get permission to pay for the it, but given the support team’s urgency
around audits, that should not be hard.

## 9. Quickly triage problems in a highly composed microservice based application  
Everything was going well and all of the sudden things stopped working for the folks in customer support.  I needed to find 
the problem.  Was it the connectivity between their app and mine, my code, the connectivity between my app and Rabbit, 
Rabbit itself, the connectivity between my app and the database, or the database itself?  Too many possibilities.  I was 
going to start working the problem all the way from beginning to end until I remembered that pretty service graph I saw on 
the kubernetes YouTube video.  I opened it up and saw that the services were all running and connected.  No alarms on failed 
systems or network errors.  But, I did see that latency for the message PUT API calls was through the roof.  I went to see 
what was wrong in my app and why it was so slow, and with a bit of looking at logs, I found that I was idling in synchronous 
transactions to the database.  I didn’t want that in the critical path for message delivery, but again, the support team did 
not want anything to happen without a log of it.  But, at least I know where the problem is.  I probably saved hours of 
poking around the system with that service graph.  Now to find out what the problem is with the database vendor.

## 10. Update deployment configuration
I called the database vendor support team and they quickly told me the problem.  Utilization is going through the roof.  I 
didn’t notice on the node.js app itself because kubernetes was auto-scaling that for me without a hitch.  (Well, lots of 
instances now, but, hey, that’s between finance and the cloud - I just need to keep support tickets down.)  But, the 
database was not scaling because we only bought the “small” configuration.  I quickly got the permission to purchase the 
upgrade, and issued the single command to update service instance with the “large” argument, and after an hour of the update 
automatically sharding and moving data around, everything was better.  It’s really nice how all that was taken care of for 
me.  Some database vendor employee must have worked really hard on that “update” service broker.  I’m glad I did not have to 
solve that.

## 11. Discover type update  
Just when I thought everything was OK, I get a message from kubernetes that my app is out of date.  How did that happen?  
Well, it turns out that the database vendor rolled out a new version of the database due to a critical security bug fix.  If 
this was a fully managed service like SFDC, they would just fix it quietly and not have to notify me, but I insisted on 
using a service that lets me have my own instance and know that it’s completely isolated from other customers, so when there 
is an update, I can choose whether or not to deploy it.  I’m now starting to think harder about the benefits of transparency 
in vendor deployments vs. the benefits of fully managed services.  But I have to admit just how cool it is that I got the 
message so quickly in such an obvious way.  They made an update and published it to the store.  Kubernetes’ record of the 
database type in the service catalog that was created when I bought the service noticed the new store entry.  I was notified 
about the update immediately since I had an instance of it running and my chat app bound to that instance.  Now to do the DB 
upgrade... 

## 12. Upgrade service instances to new service type version
Wow, the version upgrade was just as easy as the “small” to “large”!   All I needed was one update command and the database 
was quickly and non-disruptively upgraded.  (OK, I skipped the part where I tested the darn thing for a week - but you get 
the idea)  That’s just amazing.  Again, kudos to whoever is working that update service broker over at the database vendor!

## 13. Discover unapproved update to a service’s resources  
While I was on a week long vacation, the interns got a call about some customer issue and they worked it for a couple days 
and determined that the container size needed to be changed.  And they just changed it on their own since they did not want 
to bother me on vacation.  The service definition I made clearly did not make that a parameter they could change, though and 
they didn’t think to tell me about it when I got back since everything went back to normal by then.  But, I look at my 
messages and there’s this alert telling me that the service instance that’s handling all our support communications is out 
of compliance because of this setting.  I was initially upset about this, but when I looked at it in more detail, I 
discovered that the interns were spot on.  After a certain point of load (which will not stop going up), we needed to start 
using bigger containers and not just keep scaling out more and more tiny ones.  (Not only did the immediate problem go away, 
but paradoxically, the cost went down which was nice - they solved another problem I’m sure I would have had next quarter)  
But, now we need to roll this out more deliberately so that the setting sticks and that new instances are deployed properly 
as well.  It’s a good thing I got the notification or I’d have never known.  Thanks Kubernetes!

## 14. Update type, create update template, and update the service instance 
So, I do for my service what the the database people did for me.  I created a new version, tested it out, and deployed it to 
the catalog (without any marketplace in the way, since this is an internal app).  Now, I’m getting messages about my chat 
app for tech support being out of date and the tech support people are calling me asking me if I knew that the app had a new 
version (well, yeah, I know, I created it :-).  I guess those notifications really work…  But, now I’m the guy who needs to 
build the update service broker and I’m worried about that.  It’s not what the database vendor went through, thankfully, 
because the chat app is stateless.  One of the default update service brokers that come with kubernetes should just work 
with minor modifications.  I choose between blue-green (where traffic is distributed between the versions), deploying a new 
version entirely and having a cutover, and per customer traffic splitting.  Per customer won’t work because I really only 
have one customer (the support app itself) and full cutover seems risky no matter how much testing I do.  So, I select the 
blue/green update option and give it the arguments that tell it how fast to ramp from 10% to 100% on the new version.  
Everything works as expected and my service instance is upgraded to the new version and the tech support people stopped 
getting the alert about being bound to an out of date chat app.
